package transfer_out

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (s *TransferOutServiceImpl) buildBegin(
	ctx context.Context,
	token string,
	root container.InspectResponse,
	sidecars []container.InspectResponse,
) (*velez_api.Transfer_Begin, error) {
	containers, err := containerManifests(root, sidecars)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	moved := append([]container.InspectResponse{root}, sidecars...)

	items, err := s.itemManifests(ctx, moved)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	begin := &velez_api.Transfer_Begin{
		Token:      token,
		Containers: containers,
		Items:      items,
	}

	return begin, nil
}
