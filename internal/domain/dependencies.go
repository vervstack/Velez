package domain

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"

	"github.com/godverv/Velez/pkg/velez_api"
)

type Dependencies struct {
	Smerds  []SmerdDependency
	Volumes []VolumeDependency
}

type SmerdDependency struct {
	Constructor      *velez_api.CreateSmerd_Request
	RunningContainer *types.ContainerJSON
}

type VolumeDependency struct {
	Constructor    *velez_api.VolumeBindings
	ExistingVolume *volume.Volume
}
