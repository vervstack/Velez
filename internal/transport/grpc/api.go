package grpc

import (
	"github.com/godverv/Velez/pkg/velez_api"

	"github.com/docker/docker/client"
)

type Api struct {
	velez_api.UnimplementedVelezAPIServer

	version string

	dockerAPI      *client.Client
	availablePorts map[uint16]bool // mapping of ports -> is_occupied state
}
