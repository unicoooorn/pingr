package generator

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

var _ service.AlertGenerator = &llmApi{}

type llmApi struct {
	url   string
	token string
}

func NewLLMApi(url string, token string) *llmApi {
	return &llmApi{
		url:   url,
		token: token,
	}
}

func (l *llmApi) GenerateAlertMessage(
	ctx context.Context,
	subsystemInfoByName map[string]model.SubsystemInfo,
) (string, error) {
	panic("unimplemented")
}
