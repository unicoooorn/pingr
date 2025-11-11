package checker

import (
	"context"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

// статическая проверка на то что ваша структура соответствует интерфейсу
var _ service.Checker = &HttpGetChecker{}

type HttpGetChecker struct {
	Config config.Config
}

func NewHttpGetChecker(config config.Config) *HttpGetChecker {
	return &HttpGetChecker{Config: config}
}

func (r *HttpGetChecker) Check(ctx context.Context, subsystem string) (model.CheckResult, error) {
	panic("unimplemented")
}
