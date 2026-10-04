package settings_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) UpdateSettings(
	ctx context.Context,
	req *pb.UpdateSettings_Request,
) (*pb.UpdateSettings_Response, error) {
	serviceReq := domain.UpdateSettingsReq{
		IsSysboxEnabled:          req.IsSysboxEnabled,
		IsSysboxWhitelistIgnored: req.IsSysboxWhitelistIgnored,
	}

	settings, err := impl.settingsService.UpdateSettings(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error updating settings")
	}

	resp := &pb.UpdateSettings_Response{
		Settings: settingsToPb(settings),
	}

	return resp, nil
}
