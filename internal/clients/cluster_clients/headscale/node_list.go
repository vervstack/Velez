package headscale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

type listNodesResponse struct {
	Nodes []struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		GivenName   string   `json:"givenName"`
		IpAddresses []string `json:"ipAddresses"`
	} `json:"nodes"`
}

func (s *Client) ListNodes(ctx context.Context) ([]domain.VcnNode, error) {
	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := s.doAPIRequest(ctx, http.MethodGet, nodeURI, nil)
	if err != nil {
		return nil, rerrors.Wrap(err, "error executing request")
	}

	defer common.CloseWithLog(resp.Body.Close, "list nodes response body")

	if resp.StatusCode == http.StatusOK {
		nodesResp := listNodesResponse{}

		err = json.NewDecoder(resp.Body).Decode(&nodesResp)
		if err != nil {
			return nil, rerrors.Wrap(err, "error decoding response")
		}

		nodes := make([]domain.VcnNode, len(nodesResp.Nodes))
		for i, node := range nodesResp.Nodes {
			name := node.GivenName
			if name == "" {
				name = node.Name
			}

			nodes[i] = domain.VcnNode{
				Name:        name,
				IpAddresses: node.IpAddresses,
			}
		}

		return nodes, nil
	}

	return nil, rerrors.Wrap(
		user_errors.ErrHeadscaleUnexpectedStatus,
		fmt.Sprintf("listing nodes, got status: %d", resp.StatusCode),
	)
}
