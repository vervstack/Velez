package headscale

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const authHeader = "Authorization"

const (
	apiBase = "/api/v1"

	userURI       = apiBase + "/user"
	clientKeyURI  = apiBase + "/preauthkey"
	nodeURI       = apiBase + "/node"
	preAuthKeyURI = apiBase + "/preauthkey"
)

var ErrNotFound = rerrors.New("not found")

func (s *Client) doAPIRequest(ctx context.Context, method string, uri string, req any) (*http.Response, error) {
	reqEncoded, err := json.Marshal(req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling request")
	}

	r, err := http.NewRequestWithContext(ctx, method, s.headscaleApiUrl+uri, bytes.NewBuffer(reqEncoded))
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating request")
	}

	return s.execAPIRequest(r)
}

func (s *Client) execAPIRequest(r *http.Request) (*http.Response, error) {
	r.Header.Add(authHeader, "Bearer "+s.apiKey)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, rerrors.Wrap(err, "")
	}

	return resp, nil
}

func (s *Client) handleError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return rerrors.Wrap(err, "error reading response body")
	}

	return rerrors.Wrap(ErrUnexpectedStatus, resp.Status, string(body))
}

type RespError struct {
	Code    codes.Code    `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details"`
}

func (e RespError) Error() string {
	return status.Error(e.Code, e.Message).Error()
}

func (e RespError) isUniqueError() bool {
	return strings.Contains(e.Message, "UNIQUE constraint failed")
}
