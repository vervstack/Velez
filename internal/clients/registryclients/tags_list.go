package registryclients

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/utils/common"
)

type tagsListResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// TagsList lists every tag published for repo, in the order the registry
// returns them - the API gives no ordering guarantee (e.g. not necessarily
// newest-first).
func (c *Client) TagsList(ctx context.Context, repo string) ([]string, error) {
	resp, err := c.doAPIRequest(ctx, http.MethodGet, "/v2/"+repo+"/tags/list")
	if err != nil {
		return nil, rerrors.Wrap(err, "error requesting tags list")
	}
	defer common.CloseWithLog(resp.Body.Close, "close registry tags list response body")

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var out tagsListResponse

	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, rerrors.Wrap(err, "error decoding tags list response")
	}

	return out.Tags, nil
}
