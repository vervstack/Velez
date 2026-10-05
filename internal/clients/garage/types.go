package garage

import (
	"time"
)

type NodeRole struct {
	Zone     string   `json:"zone"`
	Capacity *int64   `json:"capacity"`
	Tags     []string `json:"tags"`
}

type Node struct {
	Id   string    `json:"id"`
	Addr *string   `json:"addr"`
	IsUp bool      `json:"isUp"`
	Role *NodeRole `json:"role"`
}

type ClusterStatus struct {
	LayoutVersion int64  `json:"layoutVersion"`
	Nodes         []Node `json:"nodes"`
}

type ClusterHealth struct {
	Status          string `json:"status"`
	ConnectedNodes  int    `json:"connectedNodes"`
	StorageNodesUp  int    `json:"storageNodesUp"`
	PartitionsAllOk int    `json:"partitionsAllOk"`
	Partitions      int    `json:"partitions"`
}

type NodeRoleAssign struct {
	Id       string   `json:"id"`
	Zone     string   `json:"zone"`
	Capacity *int64   `json:"capacity"`
	Tags     []string `json:"tags"`
}

type ClusterLayout struct {
	Version int64 `json:"version"`
}

type BucketKeyPerm struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	Owner bool `json:"owner"`
}

type BucketKey struct {
	AccessKeyId string        `json:"accessKeyId"`
	Name        string        `json:"name"`
	Permissions BucketKeyPerm `json:"permissions"`
}

type BucketInfo struct {
	Id            string      `json:"id"`
	Created       time.Time   `json:"created"`
	GlobalAliases []string    `json:"globalAliases"`
	Keys          []BucketKey `json:"keys"`
	Objects       uint64      `json:"objects"`
	Bytes         uint64      `json:"bytes"`
}

type BucketListItem struct {
	Id            string    `json:"id"`
	Created       time.Time `json:"created"`
	GlobalAliases []string  `json:"globalAliases"`
}

type KeyBucket struct {
	Id            string        `json:"id"`
	GlobalAliases []string      `json:"globalAliases"`
	Permissions   BucketKeyPerm `json:"permissions"`
}

type KeyInfo struct {
	AccessKeyId     string      `json:"accessKeyId"`
	Name            string      `json:"name"`
	Created         *time.Time  `json:"created"`
	SecretAccessKey *string     `json:"secretAccessKey"`
	Buckets         []KeyBucket `json:"buckets"`
}

type KeyListItem struct {
	Id      string     `json:"id"`
	Name    string     `json:"name"`
	Created *time.Time `json:"created"`
}

type updateClusterLayoutRequest struct {
	Roles []NodeRoleAssign `json:"roles"`
}

type applyClusterLayoutRequest struct {
	Version int64 `json:"version"`
}

type createBucketRequest struct {
	GlobalAlias string `json:"globalAlias"`
}

type createKeyRequest struct {
	Name string `json:"name"`
}

type bucketKeyPermChangeRequest struct {
	BucketId    string        `json:"bucketId"`
	AccessKeyId string        `json:"accessKeyId"`
	Permissions BucketKeyPerm `json:"permissions"`
}
