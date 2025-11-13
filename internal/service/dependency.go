package service

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
)

type Checker interface {
	Check(ctx context.Context, subsystem string) (model.CheckResult, error)
}

type MetricsExtractor interface {
	Extract(ctx context.Context, backend string, queries []string) (model.MetricsExtractorResult, error)
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
	Render(ctx context.Context, infos map[string]model.SubsystemInfo) ([]byte, error)
}
