package transfer_out

import (
	"slices"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func validateRequest(req domain.StartTransferReq) error {
	if req.Address == "" {
		return rerrors.Wrap(errAddressRequired)
	}

	if req.Token == "" {
		return rerrors.Wrap(errTokenRequired)
	}

	return nil
}

// selectSidecars picks the requested sidecars out of the container's own ones, in the container's order.
func selectSidecars(all []container.InspectResponse, names []string) ([]container.InspectResponse, error) {
	wanted := make(map[string]struct{}, len(names))
	for _, name := range names {
		wanted[name] = struct{}{}
	}

	selected := make([]container.InspectResponse, 0, len(wanted))

	for _, sidecar := range all {
		_, isWanted := wanted[containerName(sidecar)]
		if isWanted {
			selected = append(selected, sidecar)
		}
	}

	if len(selected) != len(wanted) {
		return nil, rerrors.Wrap(errUnknownSidecar)
	}

	return selected, nil
}

func checkTransferable(
	root container.InspectResponse,
	selected []container.InspectResponse,
	consumers []domain.TransferConsumer,
	isConsumersAcknowledged bool,
) error {
	err := checkMounts(root, selected)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if len(consumers) > 0 && !isConsumersAcknowledged {
		return rerrors.Wrap(errConsumersNotAcknowledged)
	}

	return nil
}

func checkMounts(root container.InspectResponse, selected []container.InspectResponse) error {
	if hasBindMounts(root) {
		return rerrors.Wrap(errUnsupportedMounts)
	}

	if slices.ContainsFunc(selected, hasBindMounts) {
		return rerrors.Wrap(errUnsupportedMounts)
	}

	return nil
}
