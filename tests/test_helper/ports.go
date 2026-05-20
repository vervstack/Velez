package test_helper

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
)

var (
	sharedPortManagerImpl ports.PortManager
	initPortManagerOnce   sync.Once
)

func GetSharedPortManager(t *testing.T, portsToOccupy []int) ports.PortManager {
	initPortManagerOnce.Do(func() {
		d, err := docker.NewClient(nil, "")
		require.NoError(t, err)

		usedPorts, err := d.ListOccupiedPorts(t.Context())
		require.NoError(t, err)

		sharedPortManagerImpl = ports.NewPortManager(portsToOccupy, usedPorts)
	})

	return sharedPortManagerImpl
}
