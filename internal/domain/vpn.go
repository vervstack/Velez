package domain

type VpnNamespace struct {
	Id   string
	Name string
}

type ListVpnNamespaces struct {
	Id   *string
	Name *string
}

type ConnectServiceToVpn struct {
	ServiceName string
}
