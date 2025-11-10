package service

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
)

type StatusRepo interface {
	Get(ctx context.Context, subsystem string) (model.Status, error)
	Set(ctx context.Context, subsystem string, status model.Status) error
}

type Checker interface {
	Check(ctx context.Context, subsystem string) (model.Status, error)
}
