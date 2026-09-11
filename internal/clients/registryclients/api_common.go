package registryclients

import (
	"context"
	"io"
	"net/http"

	"go.redsock.ru/rerrors"
)

//nolint:forbidigo // package-private sentinel, not shared/user-facing
var ErrUnexpectedStatus = rerrors.New("unexpected status")

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

	return rerrors.Wrap(ErrUnexpectedStatus, resp.Status, string(body))
}
