package domain

type ContainerBinding struct {
	Id            int64
	ServiceId     int64
	ServiceName   string
	NodeId        int32
	Environment   string
	ContainerName string
}
