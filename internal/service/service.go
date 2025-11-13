package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// GetStatus отдаёт закэшированный статуса подсистемы
	GetStatus(ctx context.Context, subsystem string) (model.CheckResult, error)

	// InitiateCheck инициирует проверку статуса всех подсистем
	InitiateCheck(ctx context.Context) error
}

type serviceImpl struct {
	checker              Checker
	alertSender          AlertSender
	alertGenerator       AlertGenerator
	metricsExtractor     MetricsExtractor
	infographicsRenderer InfographicsRenderer
	cfg                  config.Config
}

func New(
	checker Checker,
	alertSender AlertSender,
	alertGenerator AlertGenerator,
	metricsExtractor MetricsExtractor,
	infographicsRenderer InfographicsRenderer,
	cfg config.Config,
) *serviceImpl {
	return &serviceImpl{
		checker:              checker,
		alertSender:          alertSender,
		alertGenerator:       alertGenerator,
		metricsExtractor:     metricsExtractor,
		infographicsRenderer: infographicsRenderer,
		cfg:                  cfg,
	}
}

func (s *serviceImpl) GetStatus(ctx context.Context, subsystem string) (model.CheckResult, error) {
	if subsystem == "kak dela" {
		return model.CheckResult{Status: "ok", Details: ""}, nil
	}
	panic("unimplemented")
}

func (s *serviceImpl) InitiateCheck(ctx context.Context) error {
	statuses, err := s.check(ctx)
	if err != nil {
		return fmt.Errorf("check stage: %w", err)
	}

	unhealthyDetected := false
	for _, status := range statuses {
		if status.Status == model.PingStatusNotOk {
			unhealthyDetected = true
		}
	}

	if !unhealthyDetected {
		return nil
	}

	slog.Info("encounter unhealthy state")

	if err := s.alert(ctx, statuses); err != nil {
		return fmt.Errorf("alert stage: %w", err)
	}

	return nil
}

func (s *serviceImpl) check(ctx context.Context) (map[string]model.CheckResult, error) {
	eg, ectx := errgroup.WithContext(ctx)
	eg.SetLimit(10)

	var mu sync.Mutex
	statuses := make(map[string]model.CheckResult)

	for backend := range s.cfg.Backends {
		eg.Go(
			func() error {
				res, err := s.checker.Check(ectx, backend)
				if err != nil {
					return fmt.Errorf("check health of %s: %w", backend, err)
				}

				mu.Lock()
				statuses[backend] = res
				mu.Unlock()

				return nil
			},
		)
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("check health of the whole system: %w", err)
	}

	return statuses, nil
}

func (s *serviceImpl) alert(ctx context.Context, statuses map[string]model.CheckResult) error {
	subsystemInfoByName := make(map[string]model.SubsystemInfo)

	for backend, status := range statuses {
		metricsRes, err := s.metricsExtractor.Extract(
			ctx,
			backend,
			s.cfg.Backends[backend].MetricsQueries,
		)
		if err != nil {
			return fmt.Errorf("extarct metrics: %w", err)
		}
		subsystemInfoByName[backend] = model.SubsystemInfo{
			Check:  status,
			Metric: metricsRes,
		}
	}

	msg, err := s.alertGenerator.GenerateAlertMessage(
		ctx, subsystemInfoByName,
	)
	if err != nil {
		return fmt.Errorf("generate alert msg: %w", err)
	}

	infographic, err := s.infographicsRenderer.Render(ctx, subsystemInfoByName)
	if err != nil {
		return fmt.Errorf("render infographics: %w", err)
	}

	if err := s.alertSender.SendAlert(
		ctx,
		msg,
		infographic,
	); err != nil {
		return fmt.Errorf("send alert: %w", err)
	}

	return nil
}
