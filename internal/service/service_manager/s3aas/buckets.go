package s3aas

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/garage"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (s *Service) ListBuckets(ctx context.Context, instanceName string) ([]domain.S3Bucket, error) {
	_, client, err := s.connect(ctx, instanceName)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	buckets, err := s.listBuckets(ctx, client)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return buckets, nil
}

// Buckets without a global alias are not addressable by name and are skipped.
func (s *Service) listBuckets(ctx context.Context, client *garage.Client) ([]domain.S3Bucket, error) {
	items, err := client.ListBuckets(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing buckets")
	}

	buckets := make([]domain.S3Bucket, 0, len(items))

	for _, item := range items {
		if len(item.GlobalAliases) == 0 {
			continue
		}

		info, infoErr := client.GetBucketInfo(ctx, item.Id)
		if infoErr != nil {
			return nil, rerrors.Wrap(infoErr, "error getting bucket info")
		}

		buckets = append(buckets, bucketToDomain(info))
	}

	return buckets, nil
}

func (s *Service) CreateBucket(
	ctx context.Context, instanceName, bucketName, ownerService string,
) (domain.S3Bucket, error) {
	if bucketName == "" {
		return domain.S3Bucket{}, rerrors.Wrap(user_errors.ErrS3BucketNameRequired)
	}

	_, client, err := s.connect(ctx, instanceName)
	if err != nil {
		return domain.S3Bucket{}, rerrors.Wrap(err)
	}

	info, err := client.CreateBucket(ctx, bucketName)
	if err != nil {
		if rerrors.Is(err, garage.ErrConflict) {
			return domain.S3Bucket{}, rerrors.Wrap(user_errors.ErrS3BucketAlreadyExists)
		}

		return domain.S3Bucket{}, rerrors.Wrap(err, "error creating bucket")
	}

	if ownerService == "" {
		return bucketToDomain(info), nil
	}

	info, err = s.grantOwner(ctx, client, instanceName, info, ownerService)
	if err != nil {
		s.rollbackBucket(ctx, client, instanceName, info.Id)

		return domain.S3Bucket{}, rerrors.Wrap(err)
	}

	return bucketToDomain(info), nil
}

func (s *Service) grantOwner(
	ctx context.Context, client *garage.Client, instanceName string, info garage.BucketInfo, ownerService string,
) (garage.BucketInfo, error) {
	bucketName := firstAlias(info.GlobalAliases)

	key, err := client.CreateKey(ctx, ownerKeyName(ownerService, bucketName))
	if err != nil {
		return info, rerrors.Wrap(err, "error creating owner key")
	}

	perm := garage.BucketKeyPerm{Read: true, Write: true}

	updated, err := client.AllowBucketKey(ctx, info.Id, key.AccessKeyId, perm)
	if err != nil {
		return info, rerrors.Wrap(err, "error granting owner key access")
	}

	err = s.putKeySecret(ctx, instanceName, key)
	if err != nil {
		return info, rerrors.Wrap(err)
	}

	resourceName := instanceName + "/" + bucketName

	err = s.dataStorage.ServiceResources().UpsertResource(ctx, ownerService, resourceName, domain.S3ResourceType)
	if err != nil {
		return info, rerrors.Wrap(err, "error binding bucket to owner service")
	}

	return updated, nil
}

// Best-effort removal of everything a failed CreateBucket left behind: the
// bucket and any key named after it.
func (s *Service) rollbackBucket(ctx context.Context, client *garage.Client, instanceName, bucketId string) {
	info, err := client.GetBucketInfo(ctx, bucketId)
	if err != nil {
		log.Ctx(ctx).Warn().Str("bucket_id", bucketId).Err(err).Msg("error reading bucket for rollback")

		return
	}

	bucketName := firstAlias(info.GlobalAliases)

	keys, err := client.ListKeys(ctx)
	if err != nil {
		log.Ctx(ctx).Warn().Str("bucket_id", bucketId).Err(err).Msg("error listing keys for rollback")
	}

	err = s.deleteOwnerKeys(ctx, client, instanceName, bucketName, keys)
	if err != nil {
		log.Ctx(ctx).Warn().Str("bucket_id", bucketId).Err(err).Msg("error rolling back owner keys")
	}

	err = client.DeleteBucket(ctx, bucketId)
	if err != nil {
		log.Ctx(ctx).Warn().Str("bucket_id", bucketId).Err(err).Msg("error rolling back bucket")
	}
}

func (s *Service) DeleteBucket(ctx context.Context, instanceName, bucketName string) error {
	_, client, err := s.connect(ctx, instanceName)
	if err != nil {
		return rerrors.Wrap(err)
	}

	info, err := s.bucketByName(ctx, client, bucketName)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if info.Objects > 0 {
		return rerrors.Wrap(user_errors.ErrS3BucketNotEmpty)
	}

	keys, err := client.ListKeys(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error listing keys")
	}

	err = client.DeleteBucket(ctx, info.Id)
	if err != nil {
		if rerrors.Is(err, garage.ErrBadRequest) {
			return rerrors.Wrap(user_errors.ErrS3BucketNotEmpty)
		}

		return rerrors.Wrap(err, "error deleting bucket")
	}

	return s.deleteOwnerKeys(ctx, client, instanceName, bucketName, keys)
}

func (s *Service) deleteOwnerKeys(
	ctx context.Context, client *garage.Client, instanceName, bucketName string, keys []garage.KeyListItem,
) error {
	for _, key := range keys {
		owner := ownerFromKeyName(key.Name, bucketName)
		if owner == "" {
			continue
		}

		err := client.DeleteKey(ctx, key.Id)
		if err != nil && !rerrors.Is(err, garage.ErrNotFound) {
			return rerrors.Wrap(err, "error deleting owner key")
		}

		err = s.deleteKeySecret(ctx, instanceName, key.Id)
		if err != nil {
			return rerrors.Wrap(err)
		}

		err = s.unbindOwner(ctx, instanceName, bucketName, owner)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	return nil
}

func (s *Service) unbindOwner(ctx context.Context, instanceName, bucketName, owner string) error {
	err := s.dataStorage.ServiceResources().DeleteResource(ctx, owner, instanceName+"/"+bucketName)
	if err != nil {
		return rerrors.Wrap(err, "error deleting bucket owner binding")
	}

	return nil
}

func (s *Service) SetBucketAccess(
	ctx context.Context, instanceName string, access domain.S3BucketAccess,
) (domain.S3Bucket, error) {
	_, client, err := s.connect(ctx, instanceName)
	if err != nil {
		return domain.S3Bucket{}, rerrors.Wrap(err)
	}

	info, err := s.bucketByName(ctx, client, access.BucketName)
	if err != nil {
		return domain.S3Bucket{}, rerrors.Wrap(err)
	}

	allow, deny := permsToAllowAndDeny(access)

	if hasAnyPerm(allow) {
		_, err = client.AllowBucketKey(ctx, info.Id, access.AccessKeyId, allow)
		if err != nil {
			return domain.S3Bucket{}, rerrors.Wrap(err, "error allowing bucket access")
		}
	}

	if hasAnyPerm(deny) {
		_, err = client.DenyBucketKey(ctx, info.Id, access.AccessKeyId, deny)
		if err != nil {
			return domain.S3Bucket{}, rerrors.Wrap(err, "error denying bucket access")
		}
	}

	updated, err := client.GetBucketInfo(ctx, info.Id)
	if err != nil {
		return domain.S3Bucket{}, rerrors.Wrap(err, "error getting bucket info")
	}

	return bucketToDomain(updated), nil
}

func (s *Service) bucketByName(ctx context.Context, client *garage.Client, name string) (garage.BucketInfo, error) {
	if name == "" {
		return garage.BucketInfo{}, rerrors.Wrap(user_errors.ErrS3BucketNameRequired)
	}

	info, err := client.GetBucketInfoByAlias(ctx, name)
	if err != nil {
		if rerrors.Is(err, garage.ErrNotFound) {
			return garage.BucketInfo{}, rerrors.Wrap(user_errors.ErrS3BucketNotFound)
		}

		return garage.BucketInfo{}, rerrors.Wrap(err, "error getting bucket info")
	}

	return info, nil
}
