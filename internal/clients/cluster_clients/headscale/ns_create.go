package headscale

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/domain"
	"google.golang.org/grpc/codes"
)

func (s *Client) CreateNamespace(ctx context.Context, name string) (domain.VcnNamespace, error) {
	// region Dto
	type reqBody struct {
		Name string `json:"name"`
	}

	type response struct {
		User domain.VcnNamespace
	}
	// endregion

	r := reqBody{Name: name}

	apiResp, err := s.doAPIRequest(ctx, http.MethodPost, userURI, r)
	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err, "error creating namespace")
	}

	if apiResp.StatusCode == http.StatusOK {
		var ns response

		err = json.NewDecoder(apiResp.Body).Decode(&ns)
		_ = apiResp.Body.Close()

		if err != nil {
			return domain.VcnNamespace{}, rerrors.Wrap(err, "error decoding response")
		}

		return ns.User, nil
	}

	var e RespError

	err = json.NewDecoder(apiResp.Body).Decode(&e)
	_ = apiResp.Body.Close()

	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err, "error decoding error response")
	}

	if e.isUniqueError() {
		userErr := rerrors.NewUserError("namespace already exists", codes.AlreadyExists)

		return domain.VcnNamespace{}, rerrors.Wrap(userErr, "namespace creation failed")
	}

	return domain.VcnNamespace{}, rerrors.Wrap(e)
}
