package pg_pattern

type opt func(*Constructor)

func WithInstanceName(instanceName string) opt {
	return func(c *Constructor) {
		c.InstanceName = instanceName
	}
}

func WithExposedPort() opt {
	return func(c *Constructor) {
		c.IsPortExposed = true
	}
}

func WithPort(port uint64) opt {
	return func(c *Constructor) {
		c.ExposedToPort = &port
		c.IsPortExposed = true
	}
}

func WithPassword(pwd string) opt {
	return func(c *Constructor) {
		c.MatreshkaPg.Pwd = pwd
	}
}
