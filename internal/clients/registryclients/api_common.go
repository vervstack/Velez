package registryclients

import (
	"context"
	"io"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/user_errors"
)

func (c *Client) doAPIRequest(ctx context.Context, method, uri string) (*http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, method, c.baseURL+uri, nil)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating request")
	}

	return c.execAPIRequest(r)
}

func (c *Client) execAPIRequest(r *http.Request) (*http.Response, error) {
	if c.username != "" {
		r.SetBasicAuth(c.username, c.secret)
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, rerrors.Wrap(err, "error executing request")
	}

	return resp, nil
}

func (c *Client) handleError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return rerrors.Wrap(err, "error reading response body")
	}

	return rerrors.Wrap(user_errors.ErrRegistryUnexpectedStatus, resp.Status, string(body))
}
