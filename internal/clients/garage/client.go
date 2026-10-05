// Package garage is a client for the Garage admin API v2.
package garage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.redsock.ru/rerrors"
)

const (
	requestTimeout = 15 * time.Second
	apiPrefix      = "/v2/"
)

type Client struct {
	baseUrl    string
	adminToken string
	httpClient *http.Client
}

func New(baseUrl, adminToken string) *Client {
	return &Client{
		baseUrl:    baseUrl,
		adminToken: adminToken,
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

func (c *Client) call(ctx context.Context, method, action string, query url.Values, in, out any) error {
	target := c.baseUrl + apiPrefix + action
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	return c.send(ctx, method, target, true, in, out)
}

func (c *Client) send(ctx context.Context, method, target string, isAuthorized bool, in, out any) error {
	var body io.Reader

	if in != nil {
		payload, err := json.Marshal(in)
		if err != nil {
			return rerrors.Wrap(errEncodeRequest)
		}

		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return rerrors.Wrap(errBuildRequest)
	}

	if isAuthorized {
		req.Header.Set("Authorization", "Bearer "+c.adminToken)
	}

	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return rerrors.Wrap(errDoRequest)
	}

	defer func() { _ = resp.Body.Close() }()

	err = statusToError(resp.StatusCode)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if out == nil {
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		return rerrors.Wrap(errDecodeResponse)
	}

	return nil
}

func statusToError(code int) error {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return nil
	case code == http.StatusNotFound:
		return ErrNotFound
	case code == http.StatusBadRequest:
		return ErrBadRequest
	case code == http.StatusConflict:
		return ErrConflict
	default:
		return rerrors.Wrap(errUnexpectedStatus, fmt.Sprintf("unexpected status: %d", code))
	}
}
