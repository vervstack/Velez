package velez_api_impl

import (
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) WatchContainerLogs(
	req *velez_api.WatchContainerLogs_Request,
	stream velez_api.VelezAPI_WatchContainerLogsServer,
) error {
	ctx := stream.Context()

	_, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = impl.smerdService.WatchContainerLogs(ctx, req, stream.Send)
	if err != nil {
		return rerrors.Wrap(err, "error watching container logs")
	}

	return nil
}
