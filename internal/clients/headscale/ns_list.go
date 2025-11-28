package headscale

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Client) ListNamespaces(ctx context.Context) ([]domain.VpnNamespace, error) {
	//region Response body
	type response struct {
		Users []domain.VpnNamespace `json:"users"`
	}

	//endregion

	resp, err := s.doApiRequest(ctx, http.MethodGet, userUri, nil)
	if err != nil {
		return nil, rerrors.Wrap(err, "error executing request")
	}

	if resp.StatusCode == http.StatusOK {
		nameSpaces := response{}
		return nameSpaces.Users, json.NewDecoder(resp.Body).Decode(&nameSpaces)
	}

	return nil, rerrors.New("unexpected status")
}
