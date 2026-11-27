package headscale

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

func (s *Client) DeleteNamespace(ctx context.Context, id string) error {
	err := s.deleteNamespaceNodes(ctx, id)
	if err != nil {
		return rerrors.Wrap(err, "error deleting namespace nodes")
	}

	//nolint:bodyclose // closed inline below
	resp, err := s.doAPIRequest(ctx, http.MethodDelete, userURI+"/"+id, nil)
	if err != nil {
		return rerrors.Wrap(err, "error executing request")
	}

	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return rerrors.Wrap(user_errors.ErrHeadscaleUnexpectedStatus, "deleting namespace")
}

// Headscale refuses to delete a user that still has nodes ("user not empty:
// node(s) found"), so the namespace's nodes go first.
func (s *Client) deleteNamespaceNodes(ctx context.Context, namespaceId string) error {
	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := s.doAPIRequest(ctx, http.MethodGet, nodeURI, nil)
	if err != nil {
		return rerrors.Wrap(err, "error executing request")
	}

	defer common.CloseWithLog(resp.Body.Close, "list namespace nodes response body")

	if resp.StatusCode != http.StatusOK {
		return rerrors.Wrap(user_errors.ErrHeadscaleUnexpectedStatus, "listing namespace nodes, status: "+resp.Status)
	}

	var nodes listNamespaceNodesResponse

	err = json.NewDecoder(resp.Body).Decode(&nodes)
	if err != nil {
		return rerrors.Wrap(err, "error decoding response")
	}

	for _, node := range nodes.Nodes {
		if node.User.Id.String() != namespaceId {
			continue
		}

		err = s.deleteNode(ctx, node.Id.String())
		if err != nil {
			return rerrors.Wrap(err, "error deleting node")
		}
	}

	return nil
}

func (s *Client) deleteNode(ctx context.Context, id string) error {
	//nolint:bodyclose // closed inline below
	resp, err := s.doAPIRequest(ctx, http.MethodDelete, nodeURI+"/"+id, nil)
	if err != nil {
		return rerrors.Wrap(err, "error executing request")
	}

	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return rerrors.Wrap(user_errors.ErrHeadscaleUnexpectedStatus, "deleting node")
}

type listNamespaceNodesResponse struct {
	Nodes []struct {
		Id   json.Number `json:"id"`
		User struct {
			Id json.Number `json:"id"`
		} `json:"user"`
	} `json:"nodes"`
}
