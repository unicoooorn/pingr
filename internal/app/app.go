package app

import (
	"context"
	"fmt"
	"time"

	"github.com/unicoooorn/pingr/internal/alert/generator"
	"github.com/unicoooorn/pingr/internal/alert/sender"
	"github.com/unicoooorn/pingr/internal/checker"
	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/infographics"
	metrics_extractor "github.com/unicoooorn/pingr/internal/metrics_extactor"
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

	metricsExtractor, err := metrics_extractor.NewPrometheusMetricsExtractor(cfg.Prometheus)
	if err != nil {
		return fmt.Errorf("unable to extract metrics: %w", err)
	}

	return scheduler.NewFixedIntervalScheduler(
		service.New(
			checker.NewChecker(&cfg),
			sender.NewTgApi("todo", "todo", "todo"),
			generator.NewLLMApi(&cfg),
			metricsExtractor,
			infographics.NewImageRenderer(cfg, time.Second*10),
			cfg,
		),
		10*time.Second,
	).StartMonitoring(ctx)
}
