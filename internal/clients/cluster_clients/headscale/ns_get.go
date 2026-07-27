package headscale

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/utils/common"
)

type getNamespaceResponse struct {
	Users []domain.VcnNamespace `json:"users"`
}

func (s *Client) GetNamespace(ctx context.Context, name string) (domain.VcnNamespace, error) {
	resp, err := s.doAPIRequest(ctx, http.MethodGet, userURI+"?name="+name, nil)
	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err, "error executing request")
	}

	defer common.CloseWithLog(resp.Body.Close, "Failed to decode response for GetNamespace")

	if resp.StatusCode == http.StatusOK {
		var r getNamespaceResponse

		err = json.NewDecoder(resp.Body).Decode(&r)
		if err != nil {
			return domain.VcnNamespace{}, rerrors.Wrap(err, "error decoding response")
		}

		if len(r.Users) == 0 {
			return domain.VcnNamespace{}, nil
		}

		return r.Users[0], nil
	}

	return domain.VcnNamespace{}, nil
}
