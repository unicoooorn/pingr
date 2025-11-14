package generator

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
	"go.yaml.in/yaml/v3"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	yaBaseUrl       = "https://llm.api.cloud.yandex.net/v1"
	yaModel         = "yandexgpt"
	propmptTemplate = `You are an expert Site Reliability Engineer (SRE) and Root Cause Analysis (RCA) assistant.  
Your task is to analyze the provided data about service health, dependencies, and metrics to identify the most likely root cause of the current incident.
---
## 1. System Configuration (Service Dependency Graph in YAML)
` + "`" + "`" + "`" + `yaml
{{.ServiceDeps}}
` + "`" + "`" + "`" + `
## 2. Service Statuses
{{.ServiceStatusesTable}}
## 3. Related Metrics
The following metrics are pulled from Prometheus for each affected service.
` + "`" + "`" + "`" + `yaml
{{.ServiceMetrics}}
` + "`" + "`" + "`" + `
## 4. Task
Using the data above, perform a root cause analysis.
Your response should consist of exactly the following:
	- Root Cause Summary - the most probable service or metric degradation that triggered the incident.
	- Suggested Remediation – what actions should be taken to resolve the root issue or prevent recurrence.
## 5. Notes
- Answer concisely. No more than a few sentences for each point.
- If several services failed simultaneously, prioritize identifying the one they depend on.
- If all dependencies are healthy, analyze metrics for performance degradation or latency spikes.
- Use the dependency graph to reason causally about failure propagation.`
)

var _ service.AlertGenerator = &llmApi{}

type PromptData struct {
	ServiceDeps          string
	ServiceStatusesTable string
	ServiceMetrics       string
}

type llmApi struct {
	client *openai.Client
	model  string
	config *config.Config
}

func NewLLMApi(config *config.Config) *llmApi {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("YANDEX_CLOUD_API_KEY")),
		option.WithBaseURL(yaBaseUrl),
		option.WithHeader("OpenAI-Project", os.Getenv("YANDEX_CLOUD_FOLDER")),
	)

	return &llmApi{
		client: &client,
		model:  "gpt://" + os.Getenv("YANDEX_CLOUD_FOLDER") + "/" + yaModel + "/latest",
		config: config,
	}
}

func (l *llmApi) GenerateAlertMessage(
	ctx context.Context,
	subsystemInfoByName map[string]model.SubsystemInfo,
) (string, error) {
	prompt, err := buildPrompt(l.config, subsystemInfoByName)
	if err != nil {
		return "", fmt.Errorf("build prompt: %w", err)
	}

	slog.Info(prompt)

	resp, err := l.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: l.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Temperature: openai.Float(0.3),
	})

	if err != nil {
		return "", fmt.Errorf("failed to call LLM API: %s", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return resp.Choices[0].Message.Content, nil
}

func buildPrompt(config *config.Config, subsystemInfoByName map[string]model.SubsystemInfo) (string, error) {
	// Convert the whole config text to dependencies string
	cfgYamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("error marshalling to YAML: %v", err)
	}
	deps := string(cfgYamlBytes)

	// Build statuses table
	var sb strings.Builder
	for name, data := range subsystemInfoByName {
		sb.WriteString(fmt.Sprintf("|%-10s| %6s |\n", name, data.Check.Status))
	}
	statusTable := sb.String()

	// Build metrics YAML
	metricsMap := map[string]map[string]any{}
	for name, data := range subsystemInfoByName {
		metricsMap[name] = make(map[string]any)
		for _, metric := range data.Metric.Metrics {
			metricsMap[name][metric.Name] = metric.Value
		}
	}
	metricsYAMLBytes, err := yaml.Marshal(metricsMap)
	if err != nil {
		return "", fmt.Errorf("marshal metrics: %w", err)
	}
	metricsYAML := string(metricsYAMLBytes)

	data := PromptData{
		ServiceDeps:          deps,
		ServiceStatusesTable: statusTable,
		ServiceMetrics:       metricsYAML,
	}

	tmpl, err := template.New("prompt").Parse(propmptTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
