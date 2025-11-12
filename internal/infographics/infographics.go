package infographics

import (
	"context"

	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

var _ service.InfographicsRenderer = &ImageRenderer{}

type ImageRenderer struct {
}

func (ir *ImageRenderer) Render(ctx context.Context, infos map[string]model.SubsystemInfo) ([]byte, error) {
	panic("unimplemented")
	return nil, nil
}
