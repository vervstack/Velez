package headscale

import (
	"context"
	"io"
	"net/http"

	"go.redsock.ru/rerrors"
)

func (s *Client) DeleteNamespace(ctx context.Context, id string) error {
	resp, err := s.doApiRequest(ctx, http.MethodDelete, userUri+"/"+id, nil)
	if err != nil {
		return rerrors.Wrap(err, "error executing request")
	}

	// TODO add handling error for dangling nodes of namespace
	r, _ := io.ReadAll(resp.Body)
	_ = r

	return rerrors.New("unexpected status")
}
