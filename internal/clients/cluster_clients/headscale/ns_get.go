package headscale

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Client) GetNamespace(ctx context.Context, name string) (domain.VcnNamespace, error) {
	// region Response body
	type response struct {
		Users []domain.VcnNamespace
	}
	// endregion

	resp, err := s.doAPIRequest(ctx, http.MethodGet, userURI+"?name="+name, nil)
	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err, "error executing request")
	}

	if resp.StatusCode == http.StatusOK {
		var r response

		err = json.NewDecoder(resp.Body).Decode(&r)
		_ = resp.Body.Close()

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
