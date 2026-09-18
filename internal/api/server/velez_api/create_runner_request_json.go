package velez_api

import (
	"encoding/json"

	"go.redsock.ru/rerrors"
)

// CreateRunner_Request.ProviderConfig is a oneof (a Go interface field).
// internal/jobs persists CreateRunnerTaskPayload (which embeds this message)
// via plain encoding/json, not protojson - encoding/json can marshal
// whatever concrete type ProviderConfig holds, but can never unmarshal it
// back into the interface field, silently leaving it nil. This hand-written
// pair keeps it intact across a round trip: each oneof variant is flattened
// into its own explicit, tagged field on a shadow struct, and the real
// interface field is reassembled on unmarshal. See CLAUDE.md's "Proto oneof
// fields silently break encoding/json round-trips" note.
type createRunnerRequestJSON struct {
	Name                string        `json:"name,omitempty"`
	Scope               RunnerScope   `json:"scope,omitempty"`
	Target              string        `json:"target,omitempty"`
	Labels              []string      `json:"labels,omitempty"`
	Environment         *string       `json:"environment,omitempty"`
	DockerSocketAddress *string       `json:"docker_socket_address,omitempty"`
	Github              *GithubConfig `json:"github,omitempty"`
	Gitlab              *GitlabConfig `json:"gitlab,omitempty"`
}

func (x *CreateRunner_Request) MarshalJSON() ([]byte, error) {
	shadow := createRunnerRequestJSON{
		Name:                x.GetName(),
		Scope:               x.GetScope(),
		Target:              x.GetTarget(),
		Labels:              x.GetLabels(),
		Environment:         x.Environment,
		DockerSocketAddress: x.DockerSocketAddress,
	}

	switch cfg := x.GetProviderConfig().(type) {
	case *CreateRunner_Request_Github:
		shadow.Github = cfg.Github
	case *CreateRunner_Request_Gitlab:
		shadow.Gitlab = cfg.Gitlab
	}

	data, err := json.Marshal(shadow)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshaling create runner request")
	}

	return data, nil
}

func (x *CreateRunner_Request) UnmarshalJSON(data []byte) error {
	var shadow createRunnerRequestJSON

	err := json.Unmarshal(data, &shadow)
	if err != nil {
		return rerrors.Wrap(err, "error unmarshaling create runner request")
	}

	*x = CreateRunner_Request{
		Name:                shadow.Name,
		Scope:               shadow.Scope,
		Target:              shadow.Target,
		Labels:              shadow.Labels,
		Environment:         shadow.Environment,
		DockerSocketAddress: shadow.DockerSocketAddress,
	}

	switch {
	case shadow.Github != nil:
		x.ProviderConfig = &CreateRunner_Request_Github{Github: shadow.Github}
	case shadow.Gitlab != nil:
		x.ProviderConfig = &CreateRunner_Request_Gitlab{Gitlab: shadow.Gitlab}
	}

	return nil
}
