package repo

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

// статическая проверка на то что ваша структура соответствует интерфейсу
var _ service.StatusRepo = &InMemory{}

type InMemory struct {
}

func NewInMemory() *InMemory {
	return &InMemory{}
}

func (im *InMemory) Get(ctx context.Context, subsystem string) (model.Status, error) {
	panic("unimplemented")
}

func (im *InMemory) Set(ctx context.Context, subsystem string, status model.Status) error {
	panic("unimplemented")
}
