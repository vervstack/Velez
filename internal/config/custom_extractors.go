package config

func GetAvailablePorts(cfg Config) ([]uint16, error) {
	ap := cfg.GetEnvironment().AvailablePorts
	out := make([]uint16, 0, len(ap))
	for _, p := range ap {
		out = append(out, uint16(p))
	}

	return out, nil
}
