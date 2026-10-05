package local_state

type Network struct {
	Headscale Headscale `json:"Headscale"`
}

type Headscale struct {
	ServerUrl string `json:"ServerUrl"`
	Key       string `json:"Key"`

	// LoginServerUrl is the address tailnet nodes dial to log in; empty means
	// the default public Vcn server.
	LoginServerUrl string `json:"LoginServerUrl,omitempty"`
}
