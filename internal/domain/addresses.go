package domain

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// WebUiAddressName is the registry address name that carries a service's web ui.
const (
	WebUiAddressName = "web_ui"
)

type ServiceAddress struct {
	ServiceName string
	Name        string // which exposed port, e.g. WebUiAddressName
	Host        string // "" = unknown, client substitutes the host it reached Velez on
	Port        uint32
	Scope       velez_api.AddressScope
}

type VcnNode struct {
	Name        string // headscale node name / given name == the sidecar hostname
	IpAddresses []string
}

type AddressRebuildStatus struct {
	IsRunning  bool
	TotalSteps uint32
	DoneSteps  uint32
	LastError  string // error of the last finished rebuild, "" when it succeeded or none ran yet
}
