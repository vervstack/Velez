package settings_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListWhitelistedImages(
	ctx context.Context,
	_ *pb.ListWhitelistedImages_Request,
) (*pb.ListWhitelistedImages_Response, error) {
	entries, err := impl.settingsService.ListWhitelistedImages(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing whitelisted images")
	}

	resp := &pb.ListWhitelistedImages_Response{
		Images: make([]*pb.WhitelistedImage, 0, len(entries)),
	}

	for _, entry := range entries {
		image := &pb.WhitelistedImage{
			Id:         entry.Id,
			Repository: entry.Repository,
			IsBuiltin:  entry.IsBuiltin,
		}

		resp.Images = append(resp.Images, image)
	}

	return resp, nil
}

func (impl *Impl) RemoveWhitelistedImage(
	ctx context.Context,
	req *pb.RemoveWhitelistedImage_Request,
) (*pb.RemoveWhitelistedImage_Response, error) {
	err := impl.settingsService.RemoveWhitelistedImage(ctx, req.GetId())
	if err != nil {
		return nil, rerrors.Wrap(err, "error removing whitelisted image")
	}

	resp := &pb.RemoveWhitelistedImage_Response{}

	return resp, nil
}
