package headscale

import (
	"context"
	"net/http"
	"net/url"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/utils/common"
)

func (s *Client) RegisterNode(ctx context.Context, req domain.RegisterVcnNodeReq) error {
	query := url.Values{"user": {req.Username}, "key": {req.Key}}

	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := s.doAPIRequest(ctx, http.MethodPost, nodeURI+"/register?"+query.Encode(), nil)
	if err != nil {
		return rerrors.Wrap(err, "error registering node")
	}

	defer common.CloseWithLog(resp.Body.Close, "register node response body")

	if resp.StatusCode != http.StatusOK {
		return s.handleError(resp)
	}

	return nil
}
