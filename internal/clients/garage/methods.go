package garage

import (
	"context"
	"net/http"
	"net/url"

	"go.redsock.ru/rerrors"
)

const (
	healthPath = "/health"

	queryId              = "id"
	queryGlobalAlias     = "globalAlias"
	queryShowSecretKey   = "showSecretKey"
	showSecretKeyEnabled = "true"
)

// Health calls GET /health, which is unauthenticated; 200 means ready.
func (c *Client) Health(ctx context.Context) error {
	err := c.send(ctx, http.MethodGet, c.baseUrl+healthPath, false, nil, nil)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

// CheckAdminApi succeeds once the authenticated admin API answers. Unlike
// Health it does not need a layout.
func (c *Client) CheckAdminApi(ctx context.Context) error {
	_, err := c.GetClusterStatus(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (c *Client) GetClusterStatus(ctx context.Context) (ClusterStatus, error) {
	var out ClusterStatus

	err := c.call(ctx, http.MethodGet, "GetClusterStatus", nil, nil, &out)
	if err != nil {
		return ClusterStatus{}, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) GetClusterHealth(ctx context.Context) (ClusterHealth, error) {
	var out ClusterHealth

	err := c.call(ctx, http.MethodGet, "GetClusterHealth", nil, nil, &out)
	if err != nil {
		return ClusterHealth{}, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) GetClusterLayout(ctx context.Context) (ClusterLayout, error) {
	var out ClusterLayout

	err := c.call(ctx, http.MethodGet, "GetClusterLayout", nil, nil, &out)
	if err != nil {
		return ClusterLayout{}, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) UpdateClusterLayout(ctx context.Context, roles []NodeRoleAssign) error {
	in := updateClusterLayoutRequest{Roles: roles}

	err := c.call(ctx, http.MethodPost, "UpdateClusterLayout", nil, in, nil)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (c *Client) ApplyClusterLayout(ctx context.Context, version int64) error {
	in := applyClusterLayoutRequest{Version: version}

	err := c.call(ctx, http.MethodPost, "ApplyClusterLayout", nil, in, nil)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (c *Client) ListBuckets(ctx context.Context) ([]BucketListItem, error) {
	var out []BucketListItem

	err := c.call(ctx, http.MethodGet, "ListBuckets", nil, nil, &out)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) GetBucketInfoByAlias(ctx context.Context, alias string) (BucketInfo, error) {
	query := url.Values{queryGlobalAlias: {alias}}

	return c.getBucketInfo(ctx, query)
}

func (c *Client) GetBucketInfo(ctx context.Context, id string) (BucketInfo, error) {
	query := url.Values{queryId: {id}}

	return c.getBucketInfo(ctx, query)
}

func (c *Client) getBucketInfo(ctx context.Context, query url.Values) (BucketInfo, error) {
	var out BucketInfo

	err := c.call(ctx, http.MethodGet, "GetBucketInfo", query, nil, &out)
	if err != nil {
		return BucketInfo{}, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) CreateBucket(ctx context.Context, globalAlias string) (BucketInfo, error) {
	in := createBucketRequest{GlobalAlias: globalAlias}

	var out BucketInfo

	err := c.call(ctx, http.MethodPost, "CreateBucket", nil, in, &out)
	if err != nil {
		return BucketInfo{}, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) DeleteBucket(ctx context.Context, id string) error {
	query := url.Values{queryId: {id}}

	err := c.call(ctx, http.MethodPost, "DeleteBucket", query, nil, nil)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (c *Client) ListKeys(ctx context.Context) ([]KeyListItem, error) {
	var out []KeyListItem

	err := c.call(ctx, http.MethodGet, "ListKeys", nil, nil, &out)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) GetKeyInfo(ctx context.Context, id string, isSecretShown bool) (KeyInfo, error) {
	query := url.Values{queryId: {id}}
	if isSecretShown {
		query.Set(queryShowSecretKey, showSecretKeyEnabled)
	}

	var out KeyInfo

	err := c.call(ctx, http.MethodGet, "GetKeyInfo", query, nil, &out)
	if err != nil {
		return KeyInfo{}, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) CreateKey(ctx context.Context, name string) (KeyInfo, error) {
	in := createKeyRequest{Name: name}

	var out KeyInfo

	err := c.call(ctx, http.MethodPost, "CreateKey", nil, in, &out)
	if err != nil {
		return KeyInfo{}, rerrors.Wrap(err)
	}

	return out, nil
}

func (c *Client) DeleteKey(ctx context.Context, id string) error {
	query := url.Values{queryId: {id}}

	err := c.call(ctx, http.MethodPost, "DeleteKey", query, nil, nil)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (c *Client) AllowBucketKey(
	ctx context.Context, bucketId, accessKeyId string, permissions BucketKeyPerm,
) (BucketInfo, error) {
	return c.changeBucketKey(ctx, "AllowBucketKey", bucketId, accessKeyId, permissions)
}

func (c *Client) DenyBucketKey(
	ctx context.Context, bucketId, accessKeyId string, permissions BucketKeyPerm,
) (BucketInfo, error) {
	return c.changeBucketKey(ctx, "DenyBucketKey", bucketId, accessKeyId, permissions)
}

func (c *Client) changeBucketKey(
	ctx context.Context, action, bucketId, accessKeyId string, permissions BucketKeyPerm,
) (BucketInfo, error) {
	in := bucketKeyPermChangeRequest{
		BucketId:    bucketId,
		AccessKeyId: accessKeyId,
		Permissions: permissions,
	}

	var out BucketInfo

	err := c.call(ctx, http.MethodPost, action, nil, in, &out)
	if err != nil {
		return BucketInfo{}, rerrors.Wrap(err)
	}

	return out, nil
}
