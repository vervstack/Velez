package settings_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetSettings(
	ctx context.Context,
	_ *pb.GetSettings_Request,
) (*pb.GetSettings_Response, error) {
	settings, err := impl.settingsService.GetSettings(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting settings")
	}

	resp := &pb.GetSettings_Response{
		Settings: settingsToPb(settings),
	}

	return resp, nil
}
