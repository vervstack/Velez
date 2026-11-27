package headscale

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	defaultApiKeyLifetime = 90 * 24 * time.Hour
)

type apiKeyResponse struct {
	Id         json.Number `json:"id"`
	Prefix     string      `json:"prefix"`
	Expiration time.Time   `json:"expiration"`
	CreatedAt  time.Time   `json:"createdAt"`
	LastSeen   time.Time   `json:"lastSeen"`
}

func (s *Client) ListApiKeys(ctx context.Context) ([]domain.VcnApiKey, error) {
	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := s.doAPIRequest(ctx, http.MethodGet, apiKeyURI, nil)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing Headscale API keys")
	}

	defer common.CloseWithLog(resp.Body.Close, "list API keys response body")

	if resp.StatusCode != http.StatusOK {
		return nil, s.handleError(resp)
	}

	var result struct {
		ApiKeys []apiKeyResponse `json:"apiKeys"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, rerrors.Wrap(err, "error decoding API key list")
	}

	keys := make([]domain.VcnApiKey, len(result.ApiKeys))
	for i, key := range result.ApiKeys {
		keys[i] = domain.VcnApiKey{
			Id:         key.Id.String(),
			Prefix:     key.Prefix,
			Expiration: key.Expiration,
			CreatedAt:  key.CreatedAt,
			LastSeen:   key.LastSeen,
		}
	}

	return keys, nil
}

func (s *Client) IssueApiKey(ctx context.Context) (string, error) {
	req := struct {
		Expiration time.Time `json:"expiration"`
	}{
		Expiration: time.Now().UTC().Add(defaultApiKeyLifetime),
	}

	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := s.doAPIRequest(ctx, http.MethodPost, apiKeyURI, req)
	if err != nil {
		return "", rerrors.Wrap(err, "error issuing Headscale API key")
	}

	defer common.CloseWithLog(resp.Body.Close, "issue API key response body")

	if resp.StatusCode != http.StatusOK {
		return "", s.handleError(resp)
	}

	var result struct {
		ApiKey string `json:"apiKey"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", rerrors.Wrap(err, "error decoding issued API key")
	}

	if result.ApiKey == "" {
		return "", rerrors.Wrap(user_errors.ErrHeadscaleCantParseOutput, "Headscale API returned an empty key")
	}

	return result.ApiKey, nil
}

func (s *Client) ExpireApiKey(ctx context.Context, prefix string) error {
	req := struct {
		Prefix string `json:"prefix"`
	}{
		Prefix: prefix,
	}

	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := s.doAPIRequest(ctx, http.MethodPost, apiKeyURI+"/expire", req)
	if err != nil {
		return rerrors.Wrap(err, "error expiring Headscale API key")
	}

	defer common.CloseWithLog(resp.Body.Close, "expire API key response body")

	if resp.StatusCode != http.StatusOK {
		return s.handleError(resp)
	}

	return nil
}
