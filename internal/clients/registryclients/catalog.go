package registryclients

import (
	"context"
	"encoding/json"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	catalogURI = "/v2/_catalog"
)

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

// Catalog lists every repository name known to the registry. The v2 API has
// no search endpoint, so callers substring-filter this list themselves.
func (c *Client) Catalog(ctx context.Context) ([]string, error) {
	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := c.doAPIRequest(ctx, http.MethodGet, catalogURI)
	if err != nil {
		return nil, rerrors.Wrap(err, "error requesting catalog")
	}
	defer common.CloseWithLog(resp.Body.Close, "registry catalog response body")

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var out catalogResponse

	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, rerrors.Wrap(err, "error decoding catalog response")
	}

	return out.Repositories, nil
}
