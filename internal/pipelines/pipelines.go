package pipelines

import (
	"context"

	"go.vervstack.ru/matreshka"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Pipeliner interface {
	// LaunchSmerd - prepares environment and launches container (smerd)
	// if possible (no other container with given name already exists dead or alive)
	LaunchSmerd(request domain.LaunchSmerd) Runner[domain.LaunchSmerdResult]
	// AssembleConfig - gathers information from
	// configuration file inside image
	// and matreshka instance
	// providing updated configuration
	AssembleConfig(request *velez_api.AssembleConfig_Request) Runner[matreshka.AppConfig]
}

type Runner[T any] interface {
	Run(ctx context.Context) error
	Result() (*T, error)
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
