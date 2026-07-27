package domain

type VcnNamespace struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ListVpnNamespaces struct {
	Id   *string `json:"id"`
	Name *string `json:"name"`
}

type ConnectServiceToVcn struct {
	ServiceName string `json:"service_name"`
}
type RegisterVcnNodeReq struct {
	Key      string `json:"key"`
	Username string `json:"username"`
}

type GetVcnAuthKeyReq struct {
	NamespaceId  string `json:"namespace_id"`
	ReusableOnly bool   `json:"reusable_only"`
}

type VcnAuthKey struct {
	Key string `json:"key"`
}

type IssueClientKey struct {
	NamespaceId string `json:"namespace_id"`
	Reusable    bool   `json:"reusable"`
}

type SetupHeadscaleRequest struct {
	ExposeToPort *string `json:"expose_to_port"`
	CustomImage  *string `json:"custom_image"`
}

type SetupHeadscaleResponse struct {
	Token string `json:"token"`
}
