package headscale

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Client) GetNamespace(ctx context.Context, name string) (domain.VpnNamespace, error) {
	//region Response body
	type response struct {
		Users []domain.VpnNamespace
	}
	//endregion

	resp, err := s.doApiRequest(ctx, http.MethodGet, userUri+"?name="+name, nil)
	if err != nil {
		return domain.VpnNamespace{}, rerrors.Wrap(err, "error executing request")
	}

	if resp.StatusCode == http.StatusOK {
		var r response
		err = json.NewDecoder(resp.Body).Decode(&r)
		if err != nil {
			return domain.VpnNamespace{}, rerrors.Wrap(err, "error decoding response")
		}

		if len(r.Users) == 0 {
			return domain.VpnNamespace{}, nil
		}

		return r.Users[0], nil
	}

	return domain.VpnNamespace{}, nil
}
