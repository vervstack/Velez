package pipelines

import (
	"context"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/service"
)

type Pipeliner interface {
	LaunchSmerd(request domain.LaunchSmerd) Runner[domain.LaunchSmerdResult]
}

type Runner[T any] interface {
	Run(ctx context.Context) error
	Result() (*T, error)
}

type PipelineStep interface {
	Do(ctx context.Context) error
}

type RollbackableStep interface {
	Rollback(ctx context.Context) error
}

type pipeliner struct {
	dockerAPI   clients.Docker
	portManager clients.PortManager

	configService service.ConfigurationService
}

func NewPipeliner(dockerAPI clients.Docker, portManager clients.PortManager, configService service.ConfigurationService) Pipeliner {
	return &pipeliner{
		dockerAPI:   dockerAPI,
		portManager: portManager,

		configService: configService,
	}
}
