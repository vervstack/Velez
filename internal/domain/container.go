package domain

import (
	"github.com/docker/go-connections/nat"
)

type ContainerCreate struct {
	Name            string
	ImageName       string
	AllowDuplicates bool // flag that allows to create container of image that already running

	Volumes map[string]struct{}
	Ports   map[nat.Port][]nat.PortBinding
}

type Container struct {
	UUID      string
	Name      string
	ImageName string
}
