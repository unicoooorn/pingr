package model

type PingStatus string

const (
	PingStatusOk    PingStatus = "ok"
	PingStatusNotOk PingStatus = "not_ok"
)

type Metric struct {
	Name   string
	Value  float64
	Labels map[string]string
}

// Результат работы Checker
type CheckResult struct {
	Status  PingStatus
	Details string
}

// Результат работы  MetricsExtractor
type MetricsExtractorResult struct {
	Metrics []Metric
	Details string
}
