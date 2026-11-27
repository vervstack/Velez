package headscale

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	authHeader = "Authorization"

	apiBase = "/api/v1"

	userURI       = apiBase + "/user"
	clientKeyURI  = apiBase + "/preauthkey"
	nodeURI       = apiBase + "/node"
	preAuthKeyURI = apiBase + "/preauthkey"
	apiKeyURI     = apiBase + "/apikey"
)

var apiHttpClient = &http.Client{CheckRedirect: rejectRedirect}

func rejectRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse //nolint:gocritic // net/http requires the exact sentinel
}

func (s *Client) doAPIRequest(ctx context.Context, method string, uri string, req any) (*http.Response, error) {
	var body io.Reader

	if req != nil {
		reqEncoded, err := json.Marshal(req)
		if err != nil {
			return nil, rerrors.Wrap(err, "error marshalling request")
		}

		body = bytes.NewReader(reqEncoded)
	}

	r, err := http.NewRequestWithContext(ctx, method, s.headscaleApiUrl+uri, body)
	if err != nil {
		return nil, wrapRequestError(err, "error creating request")
	}

	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}

	return s.execAPIRequest(r)
}

func (s *Client) execAPIRequest(r *http.Request) (*http.Response, error) {
	r.Header.Set(authHeader, "Bearer "+s.apiKey)

	resp, err := apiHttpClient.Do(r)
	if err != nil {
		return nil, wrapRequestError(err, "error calling Headscale")
	}

	return resp, nil
}

func wrapRequestError(err error, message string) error {
	var urlErr *url.Error

	if errors.As(err, &urlErr) {
		// The URL of a node registration request contains its registration key.
		err = urlErr.Err
	}

	return rerrors.Wrap(err, message)
}

func (s *Client) handleError(resp *http.Response) error {
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		return rerrors.Wrap(err, "error reading response body")
	}

	return rerrors.Wrap(user_errors.ErrHeadscaleUnexpectedStatus, "status: "+resp.Status)
}

type RespError struct {
	Code    codes.Code `json:"code"`
	Message string     `json:"message"`
	Details []any      `json:"details"`
}

func (e RespError) Error() string {
	return status.Error(e.Code, e.Message).Error()
}

func (e RespError) isUniqueError() bool {
	return strings.Contains(e.Message, "UNIQUE constraint failed")
}
