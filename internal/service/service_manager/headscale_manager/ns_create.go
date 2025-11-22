package headscale_manager

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Service) CreateNamespace(ctx context.Context, name string) (domain.VpnNamespace, error) {
	type reqBody struct {
		Name string `json:"name"`
	}

	r := reqBody{Name: name}
	apiResp, err := s.doApiRequest(ctx, http.MethodPost, userUri, r)
	if err != nil {
		return domain.VpnNamespace{}, rerrors.Wrap(err, "error creating namespace")
	}

	if apiResp.StatusCode == http.StatusOK {
		var ns domain.VpnNamespace
		return ns, json.NewDecoder(apiResp.Body).Decode(&ns)
	}

	var e errorResp

	err = json.NewDecoder(apiResp.Body).Decode(&e)
	if err != nil {
		return domain.VpnNamespace{}, rerrors.Wrap(err, "error decoding error response")
	}

	if e.isUniqueError() {
		return domain.VpnNamespace{}, rerrors.NewUserError("namespace already exists", codes.AlreadyExists)
	}

	return domain.VpnNamespace{}, rerrors.Wrap(e)
}
