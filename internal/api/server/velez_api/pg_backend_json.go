package velez_api

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/encoding/protojson"
)

// Task checkpoints use encoding/json, which cannot restore a oneof interface.
func (b *PgBackend) MarshalJSON() ([]byte, error) {
	data, err := protojson.Marshal(b)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return data, nil
}

func (b *PgBackend) UnmarshalJSON(data []byte) error {
	err := protojson.Unmarshal(data, b)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}
