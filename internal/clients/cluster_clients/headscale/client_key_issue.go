package headscale

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.redsock.ru/rerrors"
)

func (s *Client) IssueClientKey(ctx context.Context, namespaceId string) (string, error) {
	// region Request body
	type reqBody struct {
		NamespaceId string    `json:"user"`
		Reusable    bool      `json:"reusable"`
		Ephemeral   bool      `json:"ephemeral"`
		Expiration  time.Time `json:"expiration"`
		AclTags     []string  `json:"aclTags"`
	}
	//endregion

	// region Response body
	type response struct {
		PreAuthKey struct {
			Key string `json:"key"`
		} `json:"preAuthKey"`
	}
	//endregion

	r := reqBody{
		NamespaceId: namespaceId,
		Expiration:  time.Now().Add(time.Hour),
	}

	apiResp, err := s.doApiRequest(ctx, http.MethodPost, clientKeyUri, r)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating namespace")
	}

	if apiResp.StatusCode == http.StatusOK {
		var resp response
		return resp.PreAuthKey.Key, json.NewDecoder(apiResp.Body).Decode(&resp)
	}

	bd, err := io.ReadAll(apiResp.Body)
	if err != nil {
		return "", rerrors.Wrap(err, "error reading response body")
	}

	return "", rerrors.New("unexpected status", string(bd), apiResp.StatusCode)
}
