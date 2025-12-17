package domain

type VcnNamespace struct {
	Id   string
	Name string
}

type ListVpnNamespaces struct {
	Id   *string
	Name *string
}

type ConnectServiceToVcn struct {
	ServiceName string
}
type RegisterVcnNodeReq struct {
	Key      string
	Username string
}

type GetVcnAuthKeyReq struct {
	NamespaceId  string
	ReusableOnly bool
}

type VcnAuthKey struct {
	Key string
}

type IssueClientKey struct {
	NamespaceId string
	Reusable    bool
}
