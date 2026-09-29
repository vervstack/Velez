package domain

// SelfNodeId is the id of the node this Velez instance runs on. deploy_watcher
// reconciles deployments of this node only.
const (
	SelfNodeId int32 = 1
)

type ContainerBinding struct {
	Id            int64
	ServiceId     int64
	ServiceName   string
	NodeId        int32
	Environment   string
	ContainerName string
}
