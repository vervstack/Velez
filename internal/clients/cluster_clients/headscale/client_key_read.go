package headscale

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

func (s *Client) GetClientAuthKey(ctx context.Context, req domain.GetVcnAuthKeyReq) (
	domain.VcnAuthKey, error,
) {
	// region Response body
	type preAuthKey struct {
		User struct {
			Id json.Number `json:"id"`
		} `json:"user"`
		Key        string     `json:"key"`
		Reusable   bool       `json:"reusable"`
		Used       bool       `json:"used"`
		Expiration *time.Time `json:"expiration"`
	}

	type response struct {
		PreAuthKeys []preAuthKey `json:"preAuthKeys"`
	}

	// endregion

	query := url.Values{"user": {req.NamespaceId}}

	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := s.doAPIRequest(ctx, http.MethodGet, preAuthKeyURI+"?"+query.Encode(), nil)
	if err != nil {
		return domain.VcnAuthKey{}, rerrors.Wrap(err, "error executing request")
	}

	defer common.CloseWithLog(resp.Body.Close, "list pre-auth keys response body")

	if resp.StatusCode != http.StatusOK {
		return domain.VcnAuthKey{}, s.handleError(resp)
	}

	var r response

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return domain.VcnAuthKey{}, rerrors.Wrap(err, "error decoding response")
	}

	now := time.Now()

	for _, key := range r.PreAuthKeys {
		// Headscale 0.28+ lists all users' keys and masks the secrets of new keys.
		if key.User.Id.String() != req.NamespaceId || key.Reusable != req.ReusableOnly ||
			key.Key == "" || strings.HasSuffix(key.Key, "-***") {
			continue
		}

		isExpired := key.Expiration != nil && !key.Expiration.After(now)
		isSpent := !key.Reusable && key.Used

		if !isExpired && !isSpent {
			return domain.VcnAuthKey{Key: key.Key}, nil
		}
	}

	return domain.VcnAuthKey{}, rerrors.Wrap(user_errors.ErrNotFound, "pre-auth key not found")
}
