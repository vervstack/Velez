package headscale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

type listNamespacesResponse struct {
	Users []domain.VcnNamespace `json:"users"`
}

func (s *Client) ListNamespaces(ctx context.Context) ([]domain.VcnNamespace, error) {
	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := s.doAPIRequest(ctx, http.MethodGet, userURI, nil)
	if err != nil {
		return nil, rerrors.Wrap(err, "error executing request")
	}

	defer common.CloseWithLog(resp.Body.Close, "list namespaces response body")

	if resp.StatusCode == http.StatusOK {
		nameSpaces := listNamespacesResponse{}

		err = json.NewDecoder(resp.Body).Decode(&nameSpaces)
		if err != nil {
			return nil, rerrors.Wrap(err, "error decoding response")
		}

		return nameSpaces.Users, nil
	}

	return nil, rerrors.Wrap(
		user_errors.ErrHeadscaleUnexpectedStatus,
		fmt.Sprintf("listing namespaces, got status: %d", resp.StatusCode),
	)
}
