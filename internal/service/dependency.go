package service

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
)

type StatusRepo interface {
	Get(ctx context.Context, subsystem string) (model.CheckResult, error)
	Set(ctx context.Context, subsystem string, status model.CheckResult) error
}

type Checker interface {
	Check(ctx context.Context, subsystem string) (model.CheckResult, error)
}

type MetricsExtractor interface {
	Extract(ctx context.Context, subsystem string) (model.MetricsExtractorResult, error)
}

type AlertGenerator interface {
	GenerateAlertMessage(
		ctx context.Context,
		subsystemInfoByName map[string]model.SubsystemInfo,
	) (string, error)
}

type AlertSender interface {
	SendAlert(
		ctx context.Context,
		alertMessage string,
		infographics []byte,
	) error

	Poll(ctx context.Context) (starts []string, stops []string, err error)
}

type InfographicsRenderer interface {
	Render(ctx context.Context, infos []model.SubsystemInfo) ([]byte, error)
}
