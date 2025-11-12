package infographics

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
)

var _ service.InfographicsRenderer = &ImageRenderer{}

type ImageRenderer struct {
}

func (ir *ImageRenderer) (ctx context.Context, infos []model.SubsystemInfo) ([]byte, error) {
	panic("unimplemented")
	return nil, nil
}
