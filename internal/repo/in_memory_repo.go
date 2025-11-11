package repo

import (
	"context"
	"errors"
	"sync"

	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

// статическая проверка на то что ваша структура соответствует интерфейсу
var _ service.StatusRepo = &InMemory{}

var ErrNotFound = errors.New("subsystem not found")

type InMemory struct {
	mtx     sync.RWMutex
	storage map[string]model.CheckResult
}

func NewInMemory() *InMemory {
	return &InMemory{
		storage: make(map[string]model.CheckResult),
	}
}

func (im *InMemory) Get(ctx context.Context, subsystem string) (model.CheckResult, error) {
	if err := ctx.Err(); err != nil {
		return model.CheckResult{}, err
	}

	im.mtx.RLock()
	defer im.mtx.RUnlock()

	st, ok := im.storage[subsystem]
	if !ok {
		return model.CheckResult{}, ErrNotFound
	}

	return st, nil
}

func (im *InMemory) Set(ctx context.Context, subsystem string, status model.CheckResult) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	im.mtx.Lock()
	defer im.mtx.Unlock()
	
	im.storage[subsystem] = status

	return nil
}
