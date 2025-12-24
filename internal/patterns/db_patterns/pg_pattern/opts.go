package pg_pattern

type Opt func(*Constructor)

func WithInstanceName(instanceName string) Opt {
	return func(c *Constructor) {
		c.InstanceName = instanceName
	}
}

func WithExposedPort() Opt {
	return func(c *Constructor) {
		c.IsPortExposed = true
	}
}

func WithPort(port uint64) Opt {
	return func(c *Constructor) {
		c.ExposedToPort = &port
		c.IsPortExposed = true
	}
}

func WithPassword(pwd string) Opt {
	return func(c *Constructor) {
		c.MatreshkaPg.Pwd = pwd
	}
}
