package pg_pattern

// Opt is a functional option for configuring a PostgreSQL container constructor.
type Opt func(*Constructor)

// WithInstanceName sets the instance name for the PostgreSQL container.
func WithInstanceName(instanceName string) Opt {
	return func(c *Constructor) {
		c.InstanceName = instanceName
	}
}

// WithExposedPort exposes the PostgreSQL port on the container.
func WithExposedPort() Opt {
	return func(c *Constructor) {
		c.IsPortExposed = true
	}
}

// WithPort exposes the PostgreSQL port to a specific host port.
func WithPort(port uint64) Opt {
	return func(c *Constructor) {
		c.ExposedToPort = &port
		c.IsPortExposed = true
	}
}

// WithPassword sets the PostgreSQL password.
func WithPassword(pwd string) Opt {
	return func(c *Constructor) {
		c.MatreshkaPg.Pwd = pwd
	}
}
