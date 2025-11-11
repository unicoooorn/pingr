package service

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
)

type Service interface {
	// GetStatus отдаёт закэшированный статуса подсистемы
	GetStatus(ctx context.Context, subsystem string) (model.CheckResult, error)

	// RefreshStatus обновляет статус подсистемы: 1) проверяет статус 2) сохраняет статус в кэш
	RefreshStatus(ctx context.Context, subsystem string) error

	// GetSubsystems возвращает список всех зарегистрированных подсистем
	GetSubsystems(ctx context.Context) ([]string, error)
}

type serviceImpl struct {
	repo    StatusRepo
	checker Checker
}

func New(
	repo StatusRepo,
	checker Checker,
) *serviceImpl {
	return &serviceImpl{
		repo:    repo,
		checker: checker,
	}
}

func (s *serviceImpl) GetStatus(ctx context.Context, subsystem string) (model.CheckResult, error) {
	if subsystem == "kak dela" {
		return model.CheckResult{Status: "ok", Details: ""}, nil
	}
	panic("unimplemented")
}

func (s *serviceImpl) RefreshStatus(ctx context.Context, subsystem string) error {
	panic("unimplemented")
}

func (s *serviceImpl) GetSubsystems(ctx context.Context) ([]string, error) {
	panic("unimplemented")
}
