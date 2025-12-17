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

	userUri       = apiBase + "/user"
	clientKeyUri  = apiBase + "/preauthkey"
	nodeUri       = apiBase + "/node"
	preAuthKeyUri = apiBase + "/preauthkey"
)

var (
	ErrNotFound = rerrors.New("not found")
)

func (s *Client) doApiRequest(ctx context.Context, method string, uri string, req any) (*http.Response, error) {
	reqEncoded, err := json.Marshal(req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling request")
	}

	r, err := http.NewRequest(method, s.headscaleApiUrl+uri, bytes.NewBuffer(reqEncoded))
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating request")
	}
	r = r.WithContext(ctx)

	return s.execApiRequest(r)
}

func (s *Client) execApiRequest(r *http.Request) (*http.Response, error) {
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

	return rerrors.New("Unexpected status.", resp.Status, string(body))
}

type errorResp struct {
	Code    codes.Code    `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details"`
}

func (e errorResp) Error() string {
	return status.Error(e.Code, e.Message).Error()
}

func (e errorResp) isUniqueError() bool {
	return strings.Contains(e.Message, "UNIQUE constraint failed")
}
