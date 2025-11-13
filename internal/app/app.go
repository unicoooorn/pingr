package app

import (
	"context"
	"time"

	"github.com/unicoooorn/pingr/internal/alert/generator"
	"github.com/unicoooorn/pingr/internal/alert/sender"
	"github.com/unicoooorn/pingr/internal/checker"
	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/scheduler"
	"github.com/unicoooorn/pingr/internal/service"
)

type Scheduler interface {
	// StartMonitoring начинает мониторить сервисы
	// Про context.Context можно читать здесь: https://habr.com/ru/companies/nixys/articles/461723/
	// В первом приближении его можно игнорировать, но оставь TODOшку :)
	StartMonitoring(ctx context.Context) error
}

func Run(ctx context.Context, cfg config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return scheduler.NewFixedIntervalScheduler(
		service.New(
			checker.NewChecker(&cfg),
			sender.NewTgApi("todo", "todo", "todo"),
			generator.NewLLMApi(&cfg),
			nil, // todo: add renderer
			nil, // todo add metrics extractor
			cfg,
		),
		10*time.Second,
	).StartMonitoring(ctx)
}
