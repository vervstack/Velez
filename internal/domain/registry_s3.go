package domain

import (
	"net/url"
	"strings"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	RegistryStorageEnv            = "REGISTRY_STORAGE"
	RegistryStorageS3Value        = "s3"
	RegistryStorageS3RegionEnv    = "REGISTRY_STORAGE_S3_REGION"
	RegistryStorageS3EndpointEnv  = "REGISTRY_STORAGE_S3_REGIONENDPOINT"
	RegistryStorageS3BucketEnv    = "REGISTRY_STORAGE_S3_BUCKET"
	RegistryStorageS3PathStyleEnv = "REGISTRY_STORAGE_S3_FORCEPATHSTYLE"
	RegistryStorageS3AccessKeyEnv = "REGISTRY_STORAGE_S3_ACCESSKEY"
	RegistryStorageS3SecretKeyEnv = "REGISTRY_STORAGE_S3_SECRETKEY"
	RegistryStorageRedirectEnv    = "REGISTRY_STORAGE_REDIRECT_DISABLE"
)

// RegistryS3Storage asks for registry blobs to live in a bucket of a Velez S3
// instance. BucketName empty means the registry instance name.
type RegistryS3Storage struct {
	InstanceName string
	BucketName   string
}

// RegistryS3Binding is what a registry's own config says about its S3 backend.
type RegistryS3Binding struct {
	InstanceName string
	BucketName   string
	AccessKeyId  string
}

// RegistryS3BindingFromEnv reads the S3 backend out of a registry's env. The
// second result is false for a volume-backed registry.
func RegistryS3BindingFromEnv(env map[string]string) (RegistryS3Binding, bool) {
	if env[RegistryStorageEnv] != RegistryStorageS3Value {
		return RegistryS3Binding{}, false
	}

	binding := RegistryS3Binding{
		BucketName:  env[RegistryStorageS3BucketEnv],
		AccessKeyId: env[RegistryStorageS3AccessKeyEnv],
	}

	endpoint, err := url.Parse(env[RegistryStorageS3EndpointEnv])
	if err == nil {
		binding.InstanceName = strings.TrimPrefix(endpoint.Hostname(), labels.S3NamePrefix)
	}

	return binding, true
}
