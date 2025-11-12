package metrics_extractor

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/unicoooorn/pingr/internal/config"
	internalModel "github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

// Статическая проверка на то что структура соответствует интерфейсу
var _ service.MetricsExtractor = &PrometheusMetricsExtractor{}

type PrometheusMetricsExtractor struct {
	api v1.API
}

func NewPrometheusMetricsExtractor(cfg config.PrometheusConfig) (*PrometheusMetricsExtractor, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("prometheus url not configured")
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	if cfg.Timeout > 0 {
		httpClient.Timeout = time.Duration(cfg.Timeout)
	}

	if len(cfg.Headers) > 0 {
		httpClient.Transport = &headerRoundTripper{
			headers: cfg.Headers,
			next:    http.DefaultTransport,
		}
	}

	client, err := api.NewClient(api.Config{
		Address: cfg.URL,
		Client:  httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}

	return &PrometheusMetricsExtractor{
		api: v1.NewAPI(client),
	}, nil
}

func (p *PrometheusMetricsExtractor) Extract(ctx context.Context, subsystem string, queries []string) (internalModel.MetricsExtractorResult, error) {
	if len(queries) == 0 {
		return internalModel.MetricsExtractorResult{
			Metrics: []internalModel.Metric{},
			Details: "no metrics queries configured",
		}, nil
	}

	var allMetrics []internalModel.Metric
	var errors []string

	for _, queryExpr := range queries {
		enhancedQuery := addServiceLabel(queryExpr, subsystem)

		metrics, err := p.queryMetrics(ctx, enhancedQuery)
		if err != nil {
			errors = append(errors, fmt.Sprintf("query %q: %v", queryExpr, err))
			continue
		}
		allMetrics = append(allMetrics, metrics...)
	}

	details := fmt.Sprintf("extracted %d metrics", len(allMetrics))
	if len(errors) > 0 {
		details = fmt.Sprintf("partial success - %s, errors: %v", details, errors)
	}

	return internalModel.MetricsExtractorResult{
		Metrics: allMetrics,
		Details: details,
	}, nil
}

func addServiceLabel(query, subsystem string) string {
	if strings.Contains(query, "service=") {
		return query
	}

	label := fmt.Sprintf(`service="%s"`, subsystem)

	// Простая эвристика: ищем первое имя метрики и добавляем label
	// Поддерживаем случаи: metric, metric{}, metric{label="value"}
	for i, r := range query {
		if r == '{' {
			inner := ""
			if i+1 < len(query) && query[i+1] != '}' {
				inner = ","
			}
			return query[:i+1] + label + inner + query[i+1:]
		}
		if !isMetricNameChar(r) && i > 0 {
			return query[:i] + "{" + label + "}" + query[i:]
		}
	}

	return query + "{" + label + "}"
}

func isMetricNameChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_' || r == ':'
}

func (p *PrometheusMetricsExtractor) queryMetrics(ctx context.Context, query string) ([]internalModel.Metric, error) {
	result, _, err := p.api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}

	return convertToMetrics(result), nil
}

func convertToMetrics(value model.Value) []internalModel.Metric {
	var metrics []internalModel.Metric

	switch v := value.(type) {
	case model.Vector:
		for _, sample := range v {
			name := string(sample.Metric["__name__"])
			if name == "" {
				name = "unnamed_metric"
			}

			labels := make(map[string]string)
			for k, v := range sample.Metric {
				if k != "__name__" {
					labels[string(k)] = string(v)
				}
			}

			metrics = append(metrics, internalModel.Metric{
				Name:   name,
				Value:  float64(sample.Value),
				Labels: labels,
			})
		}

	case *model.Scalar:
		metrics = append(metrics, internalModel.Metric{
			Name:   "scalar_result",
			Value:  float64(v.Value),
			Labels: map[string]string{},
		})
	}

	return metrics
}

type headerRoundTripper struct {
	headers map[string]string
	next    http.RoundTripper
}

func (h *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}
	return h.next.RoundTrip(req)
}
