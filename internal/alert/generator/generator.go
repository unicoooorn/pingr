package generator

import (
	"context"
	"os"

	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	yaBaseUrl       = "https://llm.api.cloud.yandex.net/v1"
	yaModel         = "yandexgpt-lite"
	propmptTemplate = `You are an expert Site Reliability Engineer (SRE) and Root Cause Analysis (RCA) assistant.  
Your task is to analyze the provided data about service health, dependencies, and metrics to identify the most likely root cause of the current incident.
---
## 1. System Configuration (Service Dependency Graph in YAML)
` + "`" + "`" + "`" + `yaml
{{.ServiceDeps}}
` + "`" + "`" + "`" + `
## 2. Incident Snapshot (Statuses and Timestamps)
| Service                   | Status | Timestamp |
| ------------------------- | ------ | --------- |
{{.ServiceStatusesTable}}
## 3. Related Metrics (last 10 minutes)
The following metrics are pulled from Prometheus for each affected service.
{{.ServiceMetrics}}
## 4. Task
Using the data above, perform a root cause analysis.
Your response must include:
	- Human-readable Root Cause Summary
	- Reasoning
	- Evidence
## 5. Notes
- If several services failed simultaneously, prioritize identifying the one they depend on.
- If all dependencies are healthy, analyze metrics for performance degradation or latency spikes.
- Use the dependency graph to reason causally about failure propagation.`
)

var _ service.AlertGenerator = &llmApi{}

type llmApi struct {
	client *openai.Client
	model  string
}

func NewLLMApi() *llmApi {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("YANDEX_CLOUD_API_KEY")),
		option.WithBaseURL(yaBaseUrl),
		option.WithHeader("OpenAI-Project", os.Getenv("YANDEX_CLOUD_FOLDER")),
	)

	return &llmApi{
		client: &client,
		model:  "gpt://" + yaModel + "/latest",
	}
}

func (l *llmApi) GenerateAlertMessage(
	ctx context.Context,
	subsystemInfoByName map[string]model.SubsystemInfo,
) (string, error) {
	panic("unimplemented")
}
