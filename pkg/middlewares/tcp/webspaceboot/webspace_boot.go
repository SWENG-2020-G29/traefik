package webspaceboot

import (
	"context"
	"fmt"
	"reflect"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tcp"
	"github.com/traefik/traefik/v2/pkg/webspace"
)

const typeName = "WebspaceBootTCP"

var handlerType = reflect.TypeOf((*tcp.Handler)(nil)).Elem()

// webspaceBoot is a middleware used to ensure a webspace is booted before forwarding a request
type webspaceBoot struct {
	config *dynamic.WebspaceBoot
	name   string

	booter *webspace.Booter
}

// New creates a new handler.
func New(ctx context.Context, next tcp.Handler, config dynamic.WebspaceBoot, name string) (tcp.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	if config.URL == "" || config.IAMToken == "" || config.UserID == 0 {
		return nil, fmt.Errorf("URL, IAM token and user ID cannot be empty")
	}

	return &webspaceBoot{
		config: &config,
		name:   name,

		booter: webspace.NewBooter(&config),
	}, nil
}

func (w *webspaceBoot) ServeTCP(conn tcp.WriteCloser) {
	logger := log.FromContext(middlewares.GetLoggerCtx(context.Background(), w.name, typeName))
	logger.Debugf("Waiting for uid %v's webspace to be started", w.config.UserID)

	addr, err := w.booter.Boot()
	if err != nil {
		logger.WithError(err).Error("Failed to ensure webspace is booted")
		conn.Close()
		return
	}

	// TODO: Allow extra configuration?
	proxy, err := tcp.NewProxy(addr, 0, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to create TCP proxy")
		conn.Close()
		return
	}

	proxy.ServeTCP(conn)
}
