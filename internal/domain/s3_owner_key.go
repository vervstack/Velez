package domain

import (
	"strings"
)

const (
	s3OwnerKeySeparator = ":"
)

func S3OwnerKeyName(owner, bucket string) string {
	return owner + s3OwnerKeySeparator + bucket
}

func ParseS3OwnerKeyName(keyName string) (owner, bucket string, isOwnerKey bool) {
	idx := strings.LastIndex(keyName, s3OwnerKeySeparator)
	if idx < 0 {
		return "", "", false
	}

	owner = keyName[:idx]
	bucket = keyName[idx+len(s3OwnerKeySeparator):]

	if owner == "" || bucket == "" {
		return "", "", false
	}

	return owner, bucket, true
}
