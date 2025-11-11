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
	_mtx    sync.RWMutex
	_storage map[string]model.Status
}

func NewInMemory() *InMemory {
	return &InMemory{
		_storage: make(map[string]model.Status),
	}
}

func (im *InMemory) Get(ctx context.Context, subsystem string) (model.Status, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	im._mtx.RLock()
	defer im._mtx.RUnlock()

	st, ok := im._storage[subsystem]
	if !ok {
		return "", ErrNotFound
	}

	return st, nil
}

func (im *InMemory) Set(ctx context.Context, subsystem string, status model.Status) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	im._mtx.Lock()
	defer im._mtx.Unlock()
	
	if im._storage == nil {
		im._storage = make(map[string]model.Status)
	}
	im._storage[subsystem] = status

	return nil
}
