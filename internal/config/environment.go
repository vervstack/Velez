package config

import (
	"time"
)

type EnvironmentConfig struct {
	AvailablePorts      []int
	CPUDefault          float64
	CustomPassToKey     string
	DisableAPISecurity  bool
	ExposeMatreshkaPort bool
	MakoshExposePort    bool
	MakoshImageName     string
	MakoshKey           string
	MakoshPort          int
	MatreshkaPort       int
	MemorySwapMb        int
	NodeMode            bool
	PortainerEnabled    bool
	RAMMbDefault        int
	ShutDownOnExit      bool
	WatchTowerEnabled   bool
	WatchTowerInterval  time.Duration
}
