// Package public_api_guard rejects request content that only server-internal code may set.
package public_api_guard

import (
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

// ValidateCreateSmerdRequest refuses a CreateSmerd request arriving through a public entry when it carries bind
// mounts, elevated access or the elevated-access allowance label. Those are stamped only by server-internal code
// (register and upgrade capture); letting a caller set them would hand out arbitrary host paths and privileges.
func ValidateCreateSmerdRequest(req *velez_api.CreateSmerd_Request) error {
	if req == nil {
		return nil
	}

	settings := req.GetSettings()

	if len(settings.GetBindMounts()) > 0 || settings.GetElevatedAccess() != nil {
		return rerrors.Wrap(user_errors.ErrReservedSettingsNotAllowed)
	}

	_, hasReservedLabel := req.GetLabels()[labels.ElevatedAccessLabel]
	if hasReservedLabel {
		return rerrors.Wrap(user_errors.ErrReservedSettingsNotAllowed)
	}

	return nil
}
