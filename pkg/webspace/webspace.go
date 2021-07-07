package webspace

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

// Booter boots a webspace and retrieves IP address + HTTP(S) ports
type Booter struct {
	config *dynamic.WebspaceBoot
	client http.Client
}

// NewBooter creates a new Booter
func NewBooter(config *dynamic.WebspaceBoot) *Booter {
	return &Booter{
		config,

		http.Client{},
	}
}

func (b *Booter) applyToken(r *http.Request) {
	r.Header.Set("Authorization", "Bearer "+b.config.IAMToken)
}

type errorRes struct {
	Message string
}

type wsConfig struct {
	StartupDelay float64 `json:"startupDelay"`
	HTTPPort     uint16  `json:"httpPort"`
}

// Boot ensures a webspace is booted and returns an IP address + HTTP(S) port
func (b *Booter) Boot() (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/internal/id:%v/ensure-started", b.config.URL, b.config.UserID), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create boot request")
	}
	b.applyToken(req)

	res, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ensure booted request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		var errRes errorRes
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return "", fmt.Errorf("ensure booted request returned non-ok status code: %v", res.StatusCode)
		}

		return "", fmt.Errorf("ensure booted error: %v", errRes.Message)
	}

	d, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to read webspace ensure booted response: %w", err)
	}
	ip := string(d)

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%v/v1/webspace/id:%v/config", b.config.URL, b.config.UserID), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create webspace config request")
	}
	b.applyToken(req)

	res, err = b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make webspace config request: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		var errRes errorRes
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return "", fmt.Errorf("webspace config request returned non-ok status code: %v", res.StatusCode)
		}

		return "", fmt.Errorf("webspace config error: %v", errRes.Message)
	}

	var conf wsConfig
	if err := json.NewDecoder(res.Body).Decode(&conf); err != nil {
		return "", fmt.Errorf("failed to decode webspace config response: %w", err)
	}

	return fmt.Sprintf("%v:%v", ip, conf.HTTPPort), nil
}
