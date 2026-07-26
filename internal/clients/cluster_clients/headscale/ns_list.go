package headscale

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Client) ListNamespaces(ctx context.Context) ([]domain.VcnNamespace, error) {
	// region Response body
	type response struct {
		Users []domain.VcnNamespace `json:"users"`
	}

	// endregion

	resp, err := s.doAPIRequest(ctx, http.MethodGet, userURI, nil)
	if err != nil {
		return nil, rerrors.Wrap(err, "error executing request")
	}

	if resp.StatusCode == http.StatusOK {
		nameSpaces := response{}
		err = json.NewDecoder(resp.Body).Decode(&nameSpaces)
		_ = resp.Body.Close()

		if err != nil {
			return nil, rerrors.Wrap(err, "error decoding response")
		}

		return nameSpaces.Users, nil
	}

	return nil, rerrors.Wrap(ErrUnexpectedStatus, "listing namespaces")
}
