package settings_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListResourceBoxes(
	ctx context.Context, _ *pb.ListResourceBoxes_Request,
) (*pb.ListResourceBoxes_Response, error) {
	boxes, err := impl.settingsService.ListResourceBoxes(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &pb.ListResourceBoxes_Response{Boxes: make([]*pb.Box, 0, len(boxes))}
	for _, box := range boxes {
		item := &pb.Box{
			Name: box.Name, Cpu: box.Cpu, RamMb: box.RamMb, DiskMb: box.DiskMb, IsBuiltin: box.IsBuiltin,
		}

		resp.Boxes = append(resp.Boxes, item)
	}

	return resp, nil
}
