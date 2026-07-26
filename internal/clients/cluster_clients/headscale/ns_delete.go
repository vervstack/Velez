package headscale

import (
	"context"
	"io"
	"net/http"

	"go.redsock.ru/rerrors"
)

func (s *Client) DeleteNamespace(ctx context.Context, id string) error {
	resp, err := s.doAPIRequest(ctx, http.MethodDelete, userURI+"/"+id, nil)
	if err != nil {
		return rerrors.Wrap(err, "error executing request")
	}

	// TODO add handling error for dangling nodes of namespace
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return rerrors.Wrap(ErrUnexpectedStatus, "deleting namespace")
}
