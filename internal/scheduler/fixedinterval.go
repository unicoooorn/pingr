package scheduler

import (
	"context"
	"fmt"

	"github.com/unicoooorn/pingr/internal/service"
)

type FixedIntervalScheduler struct {
	svc service.Service
}

func NewFixedIntervalScheduler(
	svc service.Service,
) *FixedIntervalScheduler {
	return &FixedIntervalScheduler{
		svc: svc,
	}
}

func (fis *FixedIntervalScheduler) StartMonitoring(context.Context) error {
	fmt.Println("hello world")
	panic("unimplemented")
}
