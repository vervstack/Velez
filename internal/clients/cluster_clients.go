package clients

type ClusterClients interface {
	Configurator() Configurator
}

type clusterClientsContainer struct {
	configurator Configurator
}

func NewClusterClientsContainer(
	cfg Configurator,
) ClusterClients {
	return &clusterClientsContainer{
		configurator: cfg,
	}
}

func (c *clusterClientsContainer) Configurator() Configurator {
	return c.configurator
}
