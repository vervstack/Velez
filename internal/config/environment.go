// Code generated by RedSock CLI. DO NOT EDIT.

package config

type EnvironmentConfig struct {
	AvailablePorts      []int
	CustomPassToKey     string
	DisableAPISecurity  bool
	CPUDefault          float64
	MemorySwapMb        int
	RAMMbDefault        int
	PortainerEnabled    bool
	ShutDownOnExit      bool
	WatchTowerEnabled   bool
	WatchTowerInterval  time.Duration
	ExposeMatreshkaPort bool
	MatreshkaPort       int
	NodeMode            bool
}
