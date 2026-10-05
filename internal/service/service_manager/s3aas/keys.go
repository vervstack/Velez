package s3aas

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/garage"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (s *Service) ListKeys(ctx context.Context, instanceName string) ([]domain.S3Key, error) {
	_, client, err := s.connect(ctx, instanceName)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	items, err := client.ListKeys(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing keys")
	}

	keys := make([]domain.S3Key, 0, len(items))

	for _, item := range items {
		info, infoErr := client.GetKeyInfo(ctx, item.Id, false)
		if infoErr != nil {
			return nil, rerrors.Wrap(infoErr, "error getting key info")
		}

		keys = append(keys, keyToDomain(info))
	}

	return keys, nil
}

func (s *Service) CreateKey(
	ctx context.Context, instanceName, keyName string, access []domain.S3BucketAccess,
) (domain.S3KeyCredentials, error) {
	if keyName == "" {
		return domain.S3KeyCredentials{}, rerrors.Wrap(user_errors.ErrS3KeyNameRequired)
	}

	ref, client, err := s.connect(ctx, instanceName)
	if err != nil {
		return domain.S3KeyCredentials{}, rerrors.Wrap(err)
	}

	bucketIds, err := s.resolveBucketIds(ctx, client, access)
	if err != nil {
		return domain.S3KeyCredentials{}, rerrors.Wrap(err)
	}

	key, err := client.CreateKey(ctx, keyName)
	if err != nil {
		return domain.S3KeyCredentials{}, rerrors.Wrap(err, "error creating key")
	}

	err = s.configureKey(ctx, client, instanceName, key, access, bucketIds)
	if err != nil {
		s.rollbackKey(ctx, client, instanceName, key.AccessKeyId)

		return domain.S3KeyCredentials{}, rerrors.Wrap(err)
	}

	instance := s.describe(ctx, ref)

	return s.keyCredentials(instance, key.AccessKeyId, secretOf(key)), nil
}

func (s *Service) resolveBucketIds(
	ctx context.Context, client *garage.Client, access []domain.S3BucketAccess,
) ([]string, error) {
	ids := make([]string, 0, len(access))

	for _, entry := range access {
		info, err := s.bucketByName(ctx, client, entry.BucketName)
		if err != nil {
			return nil, rerrors.Wrap(err)
		}

		ids = append(ids, info.Id)
	}

	return ids, nil
}

func (s *Service) configureKey(
	ctx context.Context,
	client *garage.Client,
	instanceName string,
	key garage.KeyInfo,
	access []domain.S3BucketAccess,
	bucketIds []string,
) error {
	for i, entry := range access {
		perm := accessToPerm(entry)
		if !hasAnyPerm(perm) {
			continue
		}

		_, err := client.AllowBucketKey(ctx, bucketIds[i], key.AccessKeyId, perm)
		if err != nil {
			return rerrors.Wrap(err, "error granting key access")
		}
	}

	return s.putKeySecret(ctx, instanceName, key)
}

func (s *Service) rollbackKey(ctx context.Context, client *garage.Client, instanceName, accessKeyId string) {
	err := client.DeleteKey(ctx, accessKeyId)
	if err != nil {
		log.Ctx(ctx).Warn().Str("access_key_id", accessKeyId).Err(err).Msg("error rolling back key")
	}

	err = s.deleteKeySecret(ctx, instanceName, accessKeyId)
	if err != nil {
		log.Ctx(ctx).Warn().Str("access_key_id", accessKeyId).Err(err).Msg("error rolling back key secret")
	}
}

func (s *Service) DeleteKey(ctx context.Context, instanceName, accessKeyId string) error {
	_, client, err := s.connect(ctx, instanceName)
	if err != nil {
		return rerrors.Wrap(err)
	}

	info, err := client.GetKeyInfo(ctx, accessKeyId, false)
	if err != nil {
		if rerrors.Is(err, garage.ErrNotFound) {
			return rerrors.Wrap(user_errors.ErrS3KeyNotFound)
		}

		return rerrors.Wrap(err, "error getting key info")
	}

	err = client.DeleteKey(ctx, accessKeyId)
	if err != nil {
		if rerrors.Is(err, garage.ErrNotFound) {
			return rerrors.Wrap(user_errors.ErrS3KeyNotFound)
		}

		return rerrors.Wrap(err, "error deleting key")
	}

	err = s.deleteKeySecret(ctx, instanceName, accessKeyId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	owner, bucketName, isOwnerKey := parseOwnerKeyName(info.Name)
	if !isOwnerKey {
		return nil
	}

	err = s.unbindOwner(ctx, instanceName, bucketName, owner)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (s *Service) GetKeyCredentials(
	ctx context.Context, instanceName, accessKeyId string,
) (domain.S3KeyCredentials, error) {
	ref, err := s.findInstance(ctx, instanceName)
	if err != nil {
		return domain.S3KeyCredentials{}, rerrors.Wrap(err)
	}

	secret, err := s.secretsStore.Get(ctx, domain.S3KeySecretRef(instanceName, accessKeyId))
	if err != nil {
		if rerrors.Is(err, user_errors.ErrSecretNotFound) {
			return domain.S3KeyCredentials{}, rerrors.Wrap(user_errors.ErrS3KeyNotFound)
		}

		return domain.S3KeyCredentials{}, rerrors.Wrap(err, "error getting key secret")
	}

	instance := s.describe(ctx, ref)

	return s.keyCredentials(instance, accessKeyId, secret), nil
}

func (s *Service) keyCredentials(instance domain.S3Instance, accessKeyId, secret string) domain.S3KeyCredentials {
	return domain.S3KeyCredentials{
		AccessKeyId:        accessKeyId,
		SecretAccessKey:    secret,
		S3Endpoint:         s3Endpoint(domain.S3ServiceName(instance.Name), instance.S3Port),
		InternalS3Endpoint: internalEndpoint(domain.S3ServiceName(instance.Name), domain.S3ApiContainerPort),
		Region:             instance.Region,
	}
}

func (s *Service) putKeySecret(ctx context.Context, instanceName string, key garage.KeyInfo) error {
	err := s.secretsStore.Put(ctx, domain.S3KeySecretRef(instanceName, key.AccessKeyId), secretOf(key))
	if err != nil {
		return rerrors.Wrap(err, "error saving key secret")
	}

	return nil
}

func (s *Service) deleteKeySecret(ctx context.Context, instanceName, accessKeyId string) error {
	err := s.secretsStore.Delete(ctx, domain.S3KeySecretRef(instanceName, accessKeyId))
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting key secret")
	}

	return nil
}

func secretOf(key garage.KeyInfo) string {
	if key.SecretAccessKey == nil {
		return ""
	}

	return *key.SecretAccessKey
}
