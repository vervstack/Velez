package e2e

import (
	"context"
	"os"
	"testing"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/config"
)

func TestMain(m *testing.M) {
	cfg, err := config.Load(defaultConfigPath)
	if err != nil {
		panic(err)
	}

	d, err := docker.NewClient(cfg.Environment.CustomLabels, "")
	if err != nil {
		panic(err)
	}

	usedPorts, err := d.ListOccupiedPorts(context.Background())
	if err != nil {
		panic(err)
	}

	sharedPortManagerImpl = ports.NewPortManager(cfg.Environment.AvailablePorts, usedPorts)

	os.Exit(m.Run())
}
