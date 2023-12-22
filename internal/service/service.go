package service

import (
	"context"

	"github.com/godverv/Velez/internal/domain"
)

type ContainerManager interface {
	Up(ctx context.Context, container domain.CreateContainer) (domain.Container, error)
}
