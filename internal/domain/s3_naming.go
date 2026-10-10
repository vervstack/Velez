package domain

import (
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	S3ApiContainerPort   = 3900
	S3RpcContainerPort   = 3901
	S3AdminContainerPort = 3903
	S3WebUiContainerPort = 3909

	S3DefaultRegion    = "garage"
	S3ResourceType     = "s3_bucket"
	S3ConfigPath       = "/etc/garage.toml"
	S3MetaPath         = "/var/lib/garage/meta"
	S3DataPath         = "/var/lib/garage/data"
	S3WebUiUsername    = "admin"
	s3SecretScope      = "s3"
	s3NetworkPrefix    = "s3-"
	s3WebUiSuffix      = "_web_ui"
	s3MetaVolumeSuffix = "-meta"
	s3DataVolumeSuffix = "-data"

	s3AdminTokenSecretKey = "admin_token"
	s3RpcSecretKey        = "rpc_secret"
	s3WebUiPasswordKey    = "web_ui_password"
	s3KeySecretKeyPrefix  = "key_"
)

func S3NetworkName(name string) string {
	return s3NetworkPrefix + name
}

func S3ServiceName(name string) string {
	return labels.S3NamePrefix + name
}

func S3WebUiServiceName(name string) string {
	return labels.S3NamePrefix + name + s3WebUiSuffix
}

func IsS3WebUiServiceNameOf(serviceName string, candidate string) bool {
	return candidate == serviceName+s3WebUiSuffix
}

func S3MetaVolumeName(name string) string {
	return name + s3MetaVolumeSuffix
}

func S3DataVolumeName(name string) string {
	return name + s3DataVolumeSuffix
}

func S3AdminTokenSecretRef(name string) SecretRef {
	return SecretRef{Scope: s3SecretScope, Owner: name, Key: s3AdminTokenSecretKey}
}

func S3RpcSecretRef(name string) SecretRef {
	return SecretRef{Scope: s3SecretScope, Owner: name, Key: s3RpcSecretKey}
}

func S3WebUiPasswordSecretRef(name string) SecretRef {
	return SecretRef{Scope: s3SecretScope, Owner: name, Key: s3WebUiPasswordKey}
}

func S3KeySecretRef(instance, accessKeyId string) SecretRef {
	return SecretRef{Scope: s3SecretScope, Owner: instance, Key: s3KeySecretKeyPrefix + accessKeyId}
}
