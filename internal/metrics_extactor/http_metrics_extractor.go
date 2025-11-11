package metrics_extractor

import (
	"context"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

// статическая проверка на то что ваша структура соответствует интерфейсу
var _ service.MetricsExtractor = &HttpMetricsExtractor{}

type HttpMetricsExtractor struct {
	Config config.Config
}

func NewHttpMetricsChecker(config config.Config) *HttpMetricsExtractor {
	return &HttpMetricsExtractor{Config: config}
}

func (r *HttpMetricsExtractor) Extract(ctx context.Context, subsystem string) (model.MetricsExtractorResult, error) {
	panic("unimplemented")
}
