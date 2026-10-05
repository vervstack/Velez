package s3aas

import (
	"strings"

	"go.vervstack.ru/Velez/internal/clients/garage"
	"go.vervstack.ru/Velez/internal/domain"
)

const (
	ownerKeySeparator = ":"
)

// A key named "<owner>:<bucket>" holding read and write on the bucket marks
// <owner> as the bucket's owner; the key name is the only record of it.
func ownerKeyName(owner, bucket string) string {
	return owner + ownerKeySeparator + bucket
}

func ownerFromKeyName(keyName, bucket string) string {
	owner, keyBucket, isOwnerKey := parseOwnerKeyName(keyName)
	if !isOwnerKey || keyBucket != bucket {
		return ""
	}

	return owner
}

func parseOwnerKeyName(keyName string) (owner, bucket string, isOwnerKey bool) {
	owner, bucket, isSeparated := cutLast(keyName, ownerKeySeparator)
	if !isSeparated || owner == "" || bucket == "" {
		return "", "", false
	}

	return owner, bucket, true
}

func cutLast(value, separator string) (before, after string, isFound bool) {
	idx := strings.LastIndex(value, separator)
	if idx < 0 {
		return value, "", false
	}

	return value[:idx], value[idx+len(separator):], true
}

func bucketToDomain(info garage.BucketInfo) domain.S3Bucket {
	name := firstAlias(info.GlobalAliases)

	bucket := domain.S3Bucket{
		Id:          info.Id,
		Name:        name,
		Access:      make([]domain.S3BucketAccess, 0, len(info.Keys)),
		ObjectCount: info.Objects,
		Bytes:       info.Bytes,
		CreatedAt:   info.Created,
	}

	for _, key := range info.Keys {
		bucket.Access = append(bucket.Access, domain.S3BucketAccess{
			BucketName:  name,
			AccessKeyId: key.AccessKeyId,
			KeyName:     key.Name,
			IsRead:      key.Permissions.Read,
			IsWrite:     key.Permissions.Write,
			IsOwner:     key.Permissions.Owner,
		})

		owner := ownerFromKeyName(key.Name, name)
		if owner != "" && key.Permissions.Read && key.Permissions.Write {
			bucket.OwnerService = owner
		}
	}

	return bucket
}

func keyToDomain(info garage.KeyInfo) domain.S3Key {
	key := domain.S3Key{
		AccessKeyId: info.AccessKeyId,
		Name:        info.Name,
		Access:      make([]domain.S3BucketAccess, 0, len(info.Buckets)),
	}

	if info.Created != nil {
		key.CreatedAt = *info.Created
	}

	for _, bucket := range info.Buckets {
		bucketName := firstAlias(bucket.GlobalAliases)

		key.Access = append(key.Access, domain.S3BucketAccess{
			BucketName:  bucketName,
			AccessKeyId: info.AccessKeyId,
			KeyName:     info.Name,
			IsRead:      bucket.Permissions.Read,
			IsWrite:     bucket.Permissions.Write,
			IsOwner:     bucket.Permissions.Owner,
		})

		owner := ownerFromKeyName(info.Name, bucketName)
		if owner != "" && bucket.Permissions.Read && bucket.Permissions.Write {
			key.OwnerService = owner
		}
	}

	return key
}

func firstAlias(aliases []string) string {
	if len(aliases) == 0 {
		return ""
	}

	return aliases[0]
}

func accessToPerm(access domain.S3BucketAccess) garage.BucketKeyPerm {
	return garage.BucketKeyPerm{
		Read:  access.IsRead,
		Write: access.IsWrite,
		Owner: access.IsOwner,
	}
}

// Garage's allow flips only the true flags and deny only the true flags, so a
// requested state needs one call of each, the deny one carrying the inverse.
func permsToAllowAndDeny(access domain.S3BucketAccess) (allow, deny garage.BucketKeyPerm) {
	allow = accessToPerm(access)
	deny = garage.BucketKeyPerm{
		Read:  !access.IsRead,
		Write: !access.IsWrite,
		Owner: !access.IsOwner,
	}

	return allow, deny
}

func hasAnyPerm(perm garage.BucketKeyPerm) bool {
	return perm.Read || perm.Write || perm.Owner
}
