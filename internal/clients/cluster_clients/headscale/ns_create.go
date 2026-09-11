package headscale

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/utils/common"
)

type createNamespaceRequest struct {
	Name string `json:"name"`
}
type createNamespaceResponse struct {
	User domain.VcnNamespace `json:"user"`
}

//nolint:forbidigo // package-private sentinel, not shared/user-facing
var errNamespaceAlreadyExists = rerrors.NewUserError("namespace already exists", codes.AlreadyExists)

func (s *Client) CreateNamespace(ctx context.Context, name string) (domain.VcnNamespace, error) {
	r := createNamespaceRequest{Name: name}

	apiResp, err := s.doAPIRequest(ctx, http.MethodPost, userURI, r)
	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err, "error creating namespace")
	}

	defer common.CloseWithLog(apiResp.Body.Close, "close response body for create namespace request")

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
	_ = apiResp.Body.Close()

	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err, "error decoding error response")
	}

	if e.isUniqueError() {
		return domain.VcnNamespace{}, rerrors.Wrap(errNamespaceAlreadyExists, "namespace creation failed")
	}

	return domain.VcnNamespace{}, rerrors.Wrap(e)
}
