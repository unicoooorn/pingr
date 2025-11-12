package metrics_extractor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicoooorn/pingr/internal/config"
)

func TestPrometheusMetricsExtractor_Extract(t *testing.T) {
	// Создаём мок Prometheus сервер
	var receivedQueries []string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Prometheus API sends query in POST form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		query := r.FormValue("query")
		receivedQueries = append(receivedQueries, query)

		var response interface{}
		switch {
		case contains(query, "up"):
			response = map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": "vector",
					"result": []map[string]interface{}{
						{
							"metric": map[string]string{
								"__name__": "up",
								"service":  "myapi",
								"job":      "prometheus",
							},
							"value": []interface{}{1699999999, "1"},
						},
					},
				},
			}
		case contains(query, "http_requests_total"):
			response = map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": "vector",
					"result": []map[string]interface{}{
						{
							"metric": map[string]string{
								"__name__": "http_requests_total",
								"service":  "myapi",
								"method":   "GET",
								"status":   "200",
							},
							"value": []interface{}{1699999999, "42"},
						},
						{
							"metric": map[string]string{
								"__name__": "http_requests_total",
								"service":  "myapi",
								"method":   "POST",
								"status":   "201",
							},
							"value": []interface{}{1699999999, "13"},
						},
					},
				},
			}
		default:
			response = map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": "vector",
					"result":     []map[string]interface{}{},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	cfg := config.Config{
		Prometheus: config.PrometheusConfig{
			URL:     mockServer.URL,
			Timeout: time.Duration(5 * time.Second),
		},
		Backends: map[string]config.BackendConfig{
			"myapi": {
				Type: "http",
				URL:  "http://example.com",
				MetricsQueries: []string{
					"up",
					"http_requests_total",
				},
			},
		},
	}

	extractor, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	require.NoError(t, err)

	result, err := extractor.Extract(context.Background(), "myapi", cfg.Backends["myapi"].MetricsQueries)
	require.NoError(t, err)

	// Проверяем что service label был автоматически добавлен во все запросы
	require.Len(t, receivedQueries, 2, "Should have received 2 queries")
	for _, q := range receivedQueries {
		assert.Contains(t, q, `service="myapi"`, "Query should contain injected service label")
	}

	// Проверяем результаты
	assert.Len(t, result.Metrics, 3) // 1 from "up" + 2 from "http_requests_total"
	assert.Contains(t, result.Details, "extracted")

	// Проверяем первую метрику (up)
	upMetric := result.Metrics[0]
	assert.Equal(t, "up", upMetric.Name)
	assert.Equal(t, 1.0, upMetric.Value)
	assert.Equal(t, "myapi", upMetric.Labels["service"])
	assert.Equal(t, "prometheus", upMetric.Labels["job"])

	// Проверяем метрики запросов
	requestMetrics := result.Metrics[1:]
	assert.Len(t, requestMetrics, 2)

	for _, m := range requestMetrics {
		assert.Equal(t, "http_requests_total", m.Name)
		assert.Equal(t, "myapi", m.Labels["service"])
		assert.Contains(t, []string{"GET", "POST"}, m.Labels["method"])
	}
}

func TestPrometheusMetricsExtractor_Extract_NoQueries(t *testing.T) {
	cfg := config.Config{
		Prometheus: config.PrometheusConfig{
			URL: "http://localhost:9090",
		},
	}

	extractor, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	require.NoError(t, err)

	result, err := extractor.Extract(context.Background(), "test", []string{})
	require.NoError(t, err)
	assert.Empty(t, result.Metrics)
	assert.Contains(t, result.Details, "no metrics queries configured")
}

func TestPrometheusMetricsExtractor_Extract_PrometheusNotConfigured(t *testing.T) {
	cfg := config.Config{
		Prometheus: config.PrometheusConfig{
			// нет URL
		},
		Backends: map[string]config.BackendConfig{
			"test": {
				MetricsQueries: []string{"up"},
			},
		},
	}

	_, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prometheus url not configured")
}

func TestPrometheusMetricsExtractor_Extract_WithURL(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result": []map[string]interface{}{
					{
						"metric": map[string]string{
							"__name__": "test_metric",
							"service":  "test",
						},
						"value": []interface{}{1699999999, "100"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	cfg := config.Config{
		Prometheus: config.PrometheusConfig{
			URL: mockServer.URL,
		},
		Backends: map[string]config.BackendConfig{
			"test": {
				MetricsQueries: []string{"test_metric"},
			},
		},
	}

	extractor, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	require.NoError(t, err)

	result, err := extractor.Extract(context.Background(), "test", cfg.Backends["test"].MetricsQueries)
	require.NoError(t, err)

	assert.Len(t, result.Metrics, 1)
	assert.Equal(t, "test_metric", result.Metrics[0].Name)
	assert.Equal(t, 100.0, result.Metrics[0].Value)
	assert.Equal(t, "test", result.Metrics[0].Labels["service"])
}

func TestPrometheusMetricsExtractor_Extract_QueryFailure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "error",
			"errorType": "bad_data",
			"error":     "invalid query",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	cfg := config.Config{
		Prometheus: config.PrometheusConfig{
			URL: mockServer.URL,
		},
		Backends: map[string]config.BackendConfig{
			"test": {
				MetricsQueries: []string{"invalid{"},
			},
		},
	}

	extractor, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	require.NoError(t, err)

	result, err := extractor.Extract(context.Background(), "test", cfg.Backends["test"].MetricsQueries)
	require.NoError(t, err)
	assert.Contains(t, result.Details, "partial success")
	assert.Empty(t, result.Metrics)
}

func TestPrometheusMetricsExtractor_InjectServiceLabel(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		subsystem string
		expected  string
	}{
		{
			name:      "simple metric",
			query:     "up",
			subsystem: "myapi",
			expected:  `up{service="myapi"}`,
		},
		{
			name:      "metric with empty selector",
			query:     "up{}",
			subsystem: "myapi",
			expected:  `up{service="myapi"}`,
		},
		{
			name:      "metric with existing labels",
			query:     "up{job='prometheus'}",
			subsystem: "myapi",
			expected:  `up{service="myapi",job='prometheus'}`,
		},
		{
			name:      "metric with service label already present",
			query:     "up{service='other'}",
			subsystem: "myapi",
			expected:  `up{service='other'}`, // не меняем
		},
		{
			name:      "metric with multiple labels",
			query:     "http_requests_total{method='GET',status='200'}",
			subsystem: "myapi",
			expected:  `http_requests_total{service="myapi",method='GET',status='200'}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addServiceLabel(tt.query, tt.subsystem)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrometheusMetricsExtractor_ConvertToMetrics_Scalar(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "scalar",
				"result":     []interface{}{1699999999, "3.14"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	cfg := config.Config{
		Prometheus: config.PrometheusConfig{
			URL: mockServer.URL,
		},
		Backends: map[string]config.BackendConfig{
			"test": {
				MetricsQueries: []string{"scalar(some_metric)"},
			},
		},
	}

	extractor, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	require.NoError(t, err)

	result, err := extractor.Extract(context.Background(), "test", cfg.Backends["test"].MetricsQueries)
	require.NoError(t, err)

	assert.Len(t, result.Metrics, 1)
	assert.Equal(t, "scalar_result", result.Metrics[0].Name)
	assert.Equal(t, 3.14, result.Metrics[0].Value)
}

func TestHeaderRoundTripper(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем наличие кастомных заголовков
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result":     []map[string]interface{}{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	cfg := config.Config{
		Prometheus: config.PrometheusConfig{
			URL: mockServer.URL,
			Headers: map[string]string{
				"Authorization": "Bearer token123",
				"Accept":        "application/json",
			},
		},
		Backends: map[string]config.BackendConfig{
			"test": {
				MetricsQueries: []string{"up"},
			},
		},
	}

	extractor, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	require.NoError(t, err)

	_, err = extractor.Extract(context.Background(), "test", cfg.Backends["test"].MetricsQueries)
	require.NoError(t, err)
}

func TestPrometheusMetricsExtractor_MultipleBackends(t *testing.T) {
	// Тестируем что разные backends получают разные service labels
	var lastQuery string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}
		lastQuery = r.FormValue("query")

		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result": []map[string]interface{}{
					{
						"metric": map[string]string{
							"__name__": "up",
						},
						"value": []interface{}{1699999999, "1"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	cfg := config.Config{
		Prometheus: config.PrometheusConfig{
			URL: mockServer.URL,
		},
		Backends: map[string]config.BackendConfig{
			"backend1": {
				MetricsQueries: []string{"up"},
			},
			"backend2": {
				MetricsQueries: []string{"up"},
			},
		},
	}

	extractor, err := NewPrometheusMetricsExtractor(cfg.Prometheus)
	require.NoError(t, err)

	// Извлекаем метрики для backend1
	_, err = extractor.Extract(context.Background(), "backend1", cfg.Backends["backend1"].MetricsQueries)
	require.NoError(t, err)
	assert.Contains(t, lastQuery, `service="backend1"`)

	// Извлекаем метрики для backend2
	_, err = extractor.Extract(context.Background(), "backend2", cfg.Backends["backend2"].MetricsQueries)
	require.NoError(t, err)
	assert.Contains(t, lastQuery, `service="backend2"`)
}

// Вспомогательная функция для проверки подстроки
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
