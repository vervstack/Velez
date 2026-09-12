package headscale

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

type createNamespaceRequest struct {
	Name string `json:"name"`
}
type createNamespaceResponse struct {
	User domain.VcnNamespace `json:"user"`
}

func (s *Client) CreateNamespace(ctx context.Context, name string) (domain.VcnNamespace, error) {
	r := createNamespaceRequest{Name: name}

	//nolint:bodyclose // closed via common.CloseWithLog below
	apiResp, err := s.doAPIRequest(ctx, http.MethodPost, userURI, r)
	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err, "error creating namespace")
	}

	defer common.CloseWithLog(apiResp.Body.Close, "create namespace response body")

	if apiResp.StatusCode == http.StatusOK {
		var ns createNamespaceResponse

		err = json.NewDecoder(apiResp.Body).Decode(&ns)
		if err != nil {
			return domain.VcnNamespace{}, rerrors.Wrap(err, "error decoding response")
		}

		return ns.User, nil
	}

	var e RespError

	err = json.NewDecoder(apiResp.Body).Decode(&e)
	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err, "error decoding error response")
	}

	if e.isUniqueError() {
		return domain.VcnNamespace{}, rerrors.Wrap(
			user_errors.ErrHeadscaleNamespaceAlreadyExists,
			"namespace creation failed",
		)
	}

	return domain.VcnNamespace{}, rerrors.Wrap(e)
}
