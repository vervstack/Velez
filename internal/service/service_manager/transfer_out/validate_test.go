package transfer_out

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

func Test_ValidateRequest_RequiresAddressAndToken(t *testing.T) {
	err := validateRequest(domain.StartTransferReq{Token: "t"})
	require.ErrorIs(t, err, errAddressRequired)

	err = validateRequest(domain.StartTransferReq{Address: "a"})
	require.ErrorIs(t, err, errTokenRequired)

	err = validateRequest(domain.StartTransferReq{Address: "a", Token: "t"})
	require.NoError(t, err)
}

func Test_SelectSidecars_RejectsForeignName(t *testing.T) {
	all := []container.InspectResponse{newInspect(agentName, "agent:1")}

	selected, err := selectSidecars(all, []string{agentName})
	require.NoError(t, err)
	require.Len(t, selected, 1)

	_, err = selectSidecars(all, []string{"stranger"})
	require.ErrorIs(t, err, errUnknownSidecar)
}

func Test_CheckTransferable_RejectsBindMountsOnRootOrSelectedSidecar(t *testing.T) {
	plain := newInspect("app", "app:1", volumeMount("data"))
	bound := newInspect(agentName, "agent:1", bindMount("/srv"))

	err := checkTransferable(plain, nil, nil, true)
	require.NoError(t, err)

	err = checkTransferable(bound, nil, nil, true)
	require.ErrorIs(t, err, errUnsupportedMounts)

	err = checkTransferable(plain, []container.InspectResponse{bound}, nil, true)
	require.ErrorIs(t, err, errUnsupportedMounts)
}

func Test_CheckTransferable_RequiresConsumersAcknowledgement(t *testing.T) {
	root := newInspect("app", "app:1")
	consumers := []domain.TransferConsumer{{ContainerName: "web", NetworkName: "net"}}

	err := checkTransferable(root, nil, consumers, false)
	require.ErrorIs(t, err, errConsumersNotAcknowledged)

	err = checkTransferable(root, nil, consumers, true)
	require.NoError(t, err)
}
