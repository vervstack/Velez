package pg_pattern

type opt func(*PostgresConstructor)

func WithInstanceName(instanceName string) opt {
	return func(c *PostgresConstructor) {
		c.InstanceName = instanceName
	}
}
