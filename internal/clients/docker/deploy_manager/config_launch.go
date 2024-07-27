package deploy_manager

import (
	"github.com/docker/docker/api/types/container"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/pkg/velez_api"
)

func getLaunchConfig(req *velez_api.CreateSmerd_Request) (cfg *container.Config) {
	cfg = &container.Config{
		Image:       req.ImageName,
		Hostname:    req.GetName(),
		Cmd:         parser.FromCommand(req.Command),
		Healthcheck: parser.FromHealthcheck(req.Healthcheck),
		Env:         make([]string, 0, len(req.Env)),
		Labels:      req.Labels,
	}

	for k, v := range req.Env {
		cfg.Env = append(cfg.Env, k+"="+v)
	}

	return cfg
}
