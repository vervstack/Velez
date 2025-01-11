package clients

type ClusterClients interface {
	ServiceDiscovery() ServiceDiscovery
	Configurator() Configurator
	ConfigurationSynchronizer()
}

type clusterClientsContainer struct {
	serviceDiscovery ServiceDiscovery
	configurator     Configurator
}

func NewClusterClientsContainer(
	sd ServiceDiscovery,
	cfg Configurator,
) ClusterClients {
	return &clusterClientsContainer{
		serviceDiscovery: sd,
		configurator:     cfg,
	}
}

func (c *clusterClientsContainer) ServiceDiscovery() ServiceDiscovery {
	return c.serviceDiscovery
}

func (c *clusterClientsContainer) Configurator() Configurator {
	return c.configurator
}
