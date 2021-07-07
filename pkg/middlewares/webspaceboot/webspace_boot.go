package webspaceboot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"unsafe"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/healthcheck"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"github.com/traefik/traefik/v2/pkg/webspace"
	"github.com/vulcand/oxy/roundrobin"
)

const typeName = "WebspaceBoot"
const errorTitle = "Webspace Boot Error"

var handlerType = reflect.TypeOf((*http.Handler)(nil)).Elem()

// webspaceBoot is a middleware used to ensure a webspace is booted before forwarding a request
type webspaceBoot struct {
	config *dynamic.WebspaceBoot
	name   string

	booter   *webspace.Booter
	balancer healthcheck.BalancerHandler
}

// New creates a new handler.
func New(ctx context.Context, next http.Handler, config dynamic.WebspaceBoot, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	if config.URL == "" || config.IAMToken == "" || config.UserID == 0 {
		return nil, fmt.Errorf("URL, IAM token and user ID cannot be empty")
	}

	var lb healthcheck.BalancerHandler
	for {
		var ok bool
		lb, ok = next.(healthcheck.BalancerHandler)
		if ok {
			break
		}

		v := reflect.ValueOf(next)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		nextValue := v.FieldByName("next")
		if nextValue.IsZero() {
			return nil, errors.New("no load balancer config found")
		}

		// We have to do it this way since the field is private
		accessibleValue := reflect.NewAt(handlerType, unsafe.Pointer(nextValue.UnsafeAddr())).Elem()
		next = accessibleValue.Interface().(http.Handler)
	}

	return &webspaceBoot{
		config: &config,
		name:   name,

		booter:   webspace.NewBooter(&config),
		balancer: lb,
	}, nil
}

func (w *webspaceBoot) GetTracingInformation() (string, ext.SpanKindEnum) {
	return w.name, tracing.SpanKindNoneEnum
}

func (w *webspaceBoot) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := log.FromContext(middlewares.GetLoggerCtx(req.Context(), w.name, typeName))
	logger.Debugf("Waiting for uid %v's webspace to be started", w.config.UserID)

	addr, err := w.booter.Boot()
	if err != nil {
		logger.WithError(err).Error("Failed to ensure webspace is booted")
		http.Error(rw, fmt.Sprintf("%v: %v", errorTitle, err), http.StatusInternalServerError)
		return
	}

	url, err := url.Parse("http://" + addr)
	if err != nil {
		logger.WithError(err).Error("Failed to create webspace backend URL")
		http.Error(rw, fmt.Sprintf("%v: Failed to create backend URL", errorTitle), http.StatusInternalServerError)
		return
	}

	servers := w.balancer.Servers()
	for _, s := range servers {
		// In the Kubernetes case we have to define a dummy service, so remove it
		w.balancer.RemoveServer(s)
	}
	w.balancer.UpsertServer(url, roundrobin.Weight(1))

	w.balancer.ServeHTTP(rw, req)
}
