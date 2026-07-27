package velez_api

import (
	"encoding/json"

	"go.redsock.ru/rerrors"
)

// CreateSmerd_Request.Config is a proto oneof (Go interface field, no json
// tag). encoding/json.Marshal happily serializes whatever concrete type it
// holds, but Unmarshal can never allocate a concrete type back into an
// interface field - a round-trip through encoding/json silently drops
// Config. protojson handles oneofs correctly, but internal/jobs persists
// TaskContext (which embeds a full CreateSmerd_Request, e.g.
// CreateSmerdTaskPayload/UpgradeSmerdTaskPayload) via plain encoding/json,
// not protojson. These hand-written methods flatten Config into two
// explicit fields so it survives that round-trip.

// createSmerdRequestJSON's Config field shadows rawRequest's promoted,
// untagged Config field of the same name - encoding/json resolves
// same-named fields by nesting depth, and this one is shallower (depth 0
// vs. depth 1 through the embedded *rawRequest), so it always wins. Its
// value is never read; it only exists to keep encoding/json from ever
// touching the real (interface-typed) Config field directly.
type createSmerdRequestJSON struct {
	//nolint:unused
	*rawCreateSmerdRequest

	//nolint:tagliatelle // must stay "Config" (exact case) to match the
	// promoted, untagged field's default JSON key - encoding/json shadows by
	// exact key match, not just nesting depth, so a differently-cased tag
	// here would silently stop shadowing and reintroduce the original bug.
	Config      json.RawMessage      `json:"Config,omitempty"`
	ConfigVerv  *MatreshkaConfigSpec `json:"configVerv,omitempty"`
	ConfigPlain *PlainConfigSpec     `json:"configPlain,omitempty"`
}

type rawCreateSmerdRequest CreateSmerd_Request

// MarshalJSON flattens Config (see the package-level comment above) so it
// survives a subsequent UnmarshalJSON round-trip.
func (x *CreateSmerd_Request) MarshalJSON() ([]byte, error) {
	aux := createSmerdRequestJSON{
		rawCreateSmerdRequest: (*rawCreateSmerdRequest)(x),
		ConfigVerv:            x.GetVerv(),
		ConfigPlain:           x.GetPlain(),
	}

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshaling CreateSmerd_Request")
	}

	return data, nil
}

// UnmarshalJSON reconstructs Config from the flattened fields MarshalJSON
// produced (see the package-level comment above).
func (x *CreateSmerd_Request) UnmarshalJSON(data []byte) error {
	aux := createSmerdRequestJSON{
		rawCreateSmerdRequest: (*rawCreateSmerdRequest)(x),
	}

	err := json.Unmarshal(data, &aux)
	if err != nil {
		return rerrors.Wrap(err, "error unmarshaling CreateSmerd_Request")
	}

	switch {
	case aux.ConfigVerv != nil:
		x.Config = &CreateSmerd_Request_Verv{Verv: aux.ConfigVerv}
	case aux.ConfigPlain != nil:
		x.Config = &CreateSmerd_Request_Plain{Plain: aux.ConfigPlain}
	}

	return nil
}
