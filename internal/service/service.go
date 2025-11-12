package service

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
)

type Service interface {
	// GetStatus отдаёт закэшированный статуса подсистемы
	GetStatus(ctx context.Context, subsystem string) (model.CheckResult, error)

	// InitiateCheck инициирует проверку статуса всех подсистем
	InitiateCheck(ctx context.Context) error
}

type serviceImpl struct {
	repo           StatusRepo
	checker        Checker
	alertSender    AlertSender
	alertGenerator AlertGenerator
}

func New(
	repo StatusRepo,
	checker Checker,
	alertSender AlertSender,
	alertGenerator AlertGenerator,
) *serviceImpl {
	return &serviceImpl{
		repo:           repo,
		checker:        checker,
		alertSender:    alertSender,
		alertGenerator: alertGenerator,
	}
}

func (s *serviceImpl) GetStatus(ctx context.Context, subsystem string) (model.CheckResult, error) {
	if subsystem == "kak dela" {
		return model.CheckResult{Status: "ok", Details: ""}, nil
	}
	panic("unimplemented")
}

func (s *serviceImpl) InitiateCheck(ctx context.Context) error {
	panic("unimplemented")
}
