package checker

import (
	"context"
	"fmt"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

// статическая проверка на то что ваша структура соответствует интерфейсу
var _ service.Checker = &HttpChecker{}

type CheckerDelegate struct {
	Config *config.Config
}

func NewChecker(config *config.Config) *CheckerDelegate {
	return &CheckerDelegate{Config: config}
}

func (r *CheckerDelegate) Check(ctx context.Context, subsystem string) (model.CheckResult, error) {
	backend_cfg, exist := r.Config.Backends[subsystem]
	if !exist {
		return model.CheckResult{Status: "not_ok'"}, fmt.Errorf("not found cfg for '%s'", subsystem)
	}
	if backend_cfg.Type == "" {
		return model.CheckResult{Status: "not_ok'"}, fmt.Errorf("type empty or not found for '%s'", subsystem)
	}
	checker, err := r.getCheckerByType(backend_cfg.Type)
	if err != nil {
		return model.CheckResult{Status: "not_ok'"}, err
	}
	return checker.Check(ctx, subsystem)
}

func (c *CheckerDelegate) getCheckerByType(checker_type string) (service.Checker, error) {
	switch checker_type {
	case "http":
		return &HttpChecker{Config: c.Config}, nil
	default:
		return nil, fmt.Errorf("unknown backend type: '%s'", checker_type)
	}
}