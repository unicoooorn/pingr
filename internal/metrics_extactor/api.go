package metrics_extractor

import (
	"context"
	"fmt"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
)

func ExtractMetrics(ctx context.Context, cfg config.Config, subsystem string) (model.MetricsExtractorResult, error) {
	backendCfg, ok := cfg.Backends[subsystem]
	if !ok {
		return model.MetricsExtractorResult{}, fmt.Errorf("backend config for subsystem %q not found", subsystem)
	}

	extractor, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	if err != nil {
		return model.MetricsExtractorResult{}, fmt.Errorf("failed to create metrics extractor: %w", err)
	}
	return extractor.Extract(ctx, subsystem, backendCfg.MetricsQueries)
}
