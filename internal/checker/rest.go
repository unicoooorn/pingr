package checker

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

// статическая проверка на то что ваша структура соответствует интерфейсу
var _ service.Checker = &Rest{}

type Rest struct {
}

func NewRest() *Rest {
	return &Rest{}
}

func (r *Rest) Check(ctx context.Context, subsystem string) (model.Status, error) {
	panic("unimplemented")
}
