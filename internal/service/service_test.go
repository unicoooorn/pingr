package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

func TestKakDela(t *testing.T) {
	svc := service.New(nil, nil)

	res, err := svc.GetStatus(context.Background(), "kak dela")

	assert.NoError(t, err)
	assert.Equal(t, model.StatusOk, res)
}
