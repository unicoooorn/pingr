package checker

import (
	"context"
	"fmt"
	"strings"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

func NewChecker(config *config.Config) service.Checker {
	return &CheckerImpl{Config: config}
}

var _ service.Checker = &CheckerImpl{}

type CheckerImpl struct {
	Config *config.Config
}

func (r *CheckerImpl) Check(ctx context.Context, subsystem string) (model.CheckResult, error) {
	subsystem_cfg, exist := r.Config.Backends[subsystem]
	if !exist {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: "not found"}, fmt.Errorf("not found cfg for '%s'", subsystem)
	}
	if subsystem_cfg.Type == "" {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: "type empty"}, fmt.Errorf("type empty or not found for '%s'", subsystem)
	}

	addr := fmt.Sprintf("%s:%d", subsystem_cfg.Host, subsystem_cfg.Port)

	switch strings.ToLower(subsystem_cfg.Type) {
	case "http":
		if subsystem_cfg.URL == "" {
			return model.CheckResult{}, fmt.Errorf("http checker: missing url")
		}
		return CheckHttpHealth(ctx, subsystem_cfg.URL, subsystem_cfg.Headers, subsystem_cfg.Timeout)
	case "grpc":
		if subsystem_cfg.URL == "" {
			return model.CheckResult{}, fmt.Errorf("postgres checker: missing url (DSN)")
		}
		return CheckGrpcHealth(ctx, addr, subsystem_cfg.Timeout)
	case "icmp":
		if subsystem_cfg.Host == "" {
			return model.CheckResult{}, fmt.Errorf("icmp checker: missing host")
		}
		return CheckIcmpHealth(ctx, subsystem_cfg.Host, subsystem_cfg.Timeout)
	case "tcp":
		if subsystem_cfg.Host == "" || subsystem_cfg.Port == 0 {
			return model.CheckResult{}, fmt.Errorf("tcp checker: missing host and/or port")
		}
		return CheckTcpHealth(ctx, subsystem_cfg.Host, subsystem_cfg.Port, subsystem_cfg.Timeout)
	case "redis":
		if subsystem_cfg.Host == "" || subsystem_cfg.Port == 0 {
			return model.CheckResult{}, fmt.Errorf("redis checker: missing host and/or port")
		}
		return CheckRedisHealth(ctx, addr, subsystem_cfg.Timeout)
	case "postgres":
		if subsystem_cfg.URL == "" {
			return model.CheckResult{}, fmt.Errorf("postgres checker: missing url (DSN)")
		}
		return CheckPostgresHealth(ctx, subsystem_cfg.URL, subsystem_cfg.Timeout)
	default:
		return model.CheckResult{}, fmt.Errorf("unknown backend type: '%s'", subsystem_cfg.Type)
	}
}
