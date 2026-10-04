package settings_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) RunSysboxSmokeTest(
	ctx context.Context,
	_ *pb.RunSysboxSmokeTest_Request,
) (*pb.RunSysboxSmokeTest_Response, error) {
	result, err := impl.settingsService.RunSysboxSmokeTest(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error running sysbox smoke test")
	}

	resp := &pb.RunSysboxSmokeTest_Response{
		IsPassed: result.IsPassed,
	}

	if result.Failure != "" {
		resp.Failure = &result.Failure
	}

	return resp, nil
}
