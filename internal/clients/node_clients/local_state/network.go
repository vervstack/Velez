package local_state

type Network struct {
	Headscale Headscale `json:"Headscale"`
}

type Headscale struct {
	ServerUrl string `json:"ServerUrl"`
	Key       string `json:"Key"`
}
