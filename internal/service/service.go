package service

import (
	"context"

	"github.com/godverv/Velez/internal/domain"
)

type ContainerManager interface {
	CreateAndRun(ctx context.Context, req domain.ContainerCreate) (domain.Container, error)
}
