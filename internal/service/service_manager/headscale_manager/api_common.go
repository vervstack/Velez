package headscale_manager

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const authHeader = "Authorization"

const userUri = "/api/v1/user"

func (s *Service) doApiRequest(ctx context.Context, method string, uri string, req any) (*http.Response, error) {
	reqEncoded, err := json.Marshal(req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling request")
	}

	r, err := http.NewRequest(method, s.headscaleApiUrl+uri, bytes.NewBuffer(reqEncoded))
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating request")
	}

	r.Header.Add(authHeader, "Bearer "+s.apiKey)

	r = r.WithContext(ctx)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, rerrors.Wrap(err, "")
	}

	return resp, nil
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
