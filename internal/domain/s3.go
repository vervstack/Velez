package domain

import (
	"time"
)

type S3Instance struct {
	Name              string
	S3Port            uint32
	WebUiPort         uint32
	ReplicationFactor uint32
	Region            string
	Environment       string
	RemoteHost        string
	PublishedHost     string
	Status            string
	CreatedAt         time.Time
}

type S3BucketAccess struct {
	BucketName  string
	AccessKeyId string
	KeyName     string
	IsRead      bool
	IsWrite     bool
	IsOwner     bool
}

type S3Bucket struct {
	Id           string
	Name         string
	Access       []S3BucketAccess
	ObjectCount  uint64
	Bytes        uint64
	OwnerService string
	CreatedAt    time.Time
}

type S3Key struct {
	AccessKeyId  string
	Name         string
	Access       []S3BucketAccess
	OwnerService string
	CreatedAt    time.Time
}

// S3InstanceCredentials is the result of S3Service.GetInstanceCredentials.
// The web UI fields are empty when the instance has no web UI sidecar.
type S3InstanceCredentials struct {
	AdminToken         string
	S3Endpoint         string
	InternalS3Endpoint string
	Region             string
	WebUiUrl           string
	WebUiUsername      string
	WebUiPassword      string
}

type S3KeyCredentials struct {
	AccessKeyId        string
	SecretAccessKey    string
	S3Endpoint         string
	InternalS3Endpoint string
	Region             string
}
