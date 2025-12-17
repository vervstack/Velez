package headscale

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Client) GetClientAuthKey(ctx context.Context, req domain.GetVcnAuthKeyReq) (
	domain.VcnAuthKey, error) {

	//region Response body
	type preAuthKey struct {
		Id         string    `json:"id"`
		Key        string    `json:"key"`
		Reusable   bool      `json:"reusable"`
		Ephemeral  bool      `json:"ephemeral"`
		Used       bool      `json:"used"`
		Expiration time.Time `json:"expiration"`
		CreatedAt  time.Time `json:"createdAt"`
		//AclTags    []interface{} `json:"aclTags"`
	}
	type response struct {
		PreAuthKeys []preAuthKey `json:"preAuthKeys"`
	}
	//endregion

	resp, err := s.doApiRequest(ctx, http.MethodGet, preAuthKeyUri+"?user="+req.NamespaceId, nil)
	if err != nil {
		return domain.VcnAuthKey{}, rerrors.Wrap(err, "error executing request")
	}

	if resp.StatusCode == http.StatusOK {
		var r response
		err = json.NewDecoder(resp.Body).Decode(&r)
		if err != nil {
			return domain.VcnAuthKey{}, rerrors.Wrap(err, "error decoding response")
		}

		if len(r.PreAuthKeys) == 0 {
			return domain.VcnAuthKey{}, rerrors.Wrap(ErrNotFound, "no preAuthKeys found")
		}

		for _, key := range r.PreAuthKeys {
			if key.Reusable == req.ReusableOnly {
				return domain.VcnAuthKey{
					Key: r.PreAuthKeys[0].Key,
				}, nil
			}
		}

		return domain.VcnAuthKey{}, rerrors.Wrap(ErrNotFound, "preAuthKey not found")
	}

	return domain.VcnAuthKey{}, s.handleError(resp)
}

type T struct {
	PreAuthKeys []struct {
		Id         string        `json:"id"`
		Key        string        `json:"key"`
		Reusable   bool          `json:"reusable"`
		Ephemeral  bool          `json:"ephemeral"`
		Used       bool          `json:"used"`
		Expiration time.Time     `json:"expiration"`
		CreatedAt  time.Time     `json:"createdAt"`
		AclTags    []interface{} `json:"aclTags"`
	} `json:"preAuthKeys"`
}
