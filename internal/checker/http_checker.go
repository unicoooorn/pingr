package checker

import (
	"context"
	"fmt"
	"strings"
	"net/http"
	"time"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

// статическая проверка на то что ваша структура соответствует интерфейсу
var _ service.Checker = &HttpChecker{}

type HttpChecker struct {
	Config *config.Config
}

func NewHttpChecker(config *config.Config) *HttpChecker {
	return &HttpChecker{Config: config}
}

func (r *HttpChecker) Check(ctx context.Context, subsystem string) (model.CheckResult, error) {
	backendCfg, err := r.validateConfig(subsystem)
	if err != nil {
		return model.CheckResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", backendCfg.URL, nil)
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}, err
	}
	for k, v := range backendCfg.Headers {
		req.Header.Set(k, v)
	}

	timeout := time.Duration(backendCfg.Timeout) * time.Second
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return model.CheckResult{
			Status:  model.PingStatusNotOk,
			Details: err.Error(),
		}, err
	}
	defer resp.Body.Close()

	var status model.PingStatus
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = model.PingStatusOk
	} else {
		status = model.PingStatusNotOk
	}

	return model.CheckResult{
		Status:  status,
		Details: fmt.Sprintf("http status code: %d", resp.StatusCode),
	}, nil
}

func (c *HttpChecker) validateConfig(subsystem string) (config.BackendConfig, error) {
	backendCfg, exist := c.Config.Backends[subsystem]
	if !exist {
		return config.BackendConfig{}, fmt.Errorf("config for %s not found for HttpChecker", subsystem)
	}
	if strings.ToLower(backendCfg.Type) != "http" {
		return config.BackendConfig{}, fmt.Errorf("unexpected type of %s for HttpChecker", subsystem)
	}
	if backendCfg.URL == "" {
		return config.BackendConfig{}, fmt.Errorf("empty url in backend config for %s", subsystem)
	}
	return backendCfg, nil
}
