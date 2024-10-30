package config

func GetAvailablePorts(cfg Config) ([]uint32, error) {
	ap := cfg.Environment.AvailablePorts
	out := make([]uint32, 0, len(ap))
	for _, p := range ap {
		out = append(out, uint32(p))
	}

	return out, nil
}
