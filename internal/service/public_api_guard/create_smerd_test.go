package public_api_guard

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func newPlainRequest() *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:      "web",
		ImageName: "nginx:1",
		Labels:    map[string]string{"team": "core"},
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{{ServicePortNumber: 80}},
		},
	}
}

func Test_ValidateCreateSmerdRequest_PlainRequestPasses(t *testing.T) {
	err := ValidateCreateSmerdRequest(newPlainRequest())
	require.NoError(t, err)
}

func Test_ValidateCreateSmerdRequest_NilAndEmptyPass(t *testing.T) {
	require.NoError(t, ValidateCreateSmerdRequest(nil))
	require.NoError(t, ValidateCreateSmerdRequest(&velez_api.CreateSmerd_Request{}))
}

func Test_ValidateCreateSmerdRequest_BindMountsRefused(t *testing.T) {
	req := newPlainRequest()

	req.Settings.BindMounts = []*velez_api.BindMount{{}}

	err := ValidateCreateSmerdRequest(req)
	require.ErrorIs(t, err, user_errors.ErrReservedSettingsNotAllowed)
}

func Test_ValidateCreateSmerdRequest_ElevatedAccessRefused(t *testing.T) {
	req := newPlainRequest()

	req.Settings.ElevatedAccess = &velez_api.ElevatedAccess{}

	err := ValidateCreateSmerdRequest(req)
	require.ErrorIs(t, err, user_errors.ErrReservedSettingsNotAllowed)
}

func Test_ValidateCreateSmerdRequest_ReservedLabelRefused(t *testing.T) {
	req := newPlainRequest()

	req.Labels[labels.ElevatedAccessLabel] = labels.ElevatedAccessAllowedValue

	err := ValidateCreateSmerdRequest(req)
	require.ErrorIs(t, err, user_errors.ErrReservedSettingsNotAllowed)
}
