package common

import (
	"time"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func ToNodeList(info []domain.NodeBaseInfo) []*pb.NodeBaseInfo {
	nodes := make([]*pb.NodeBaseInfo, 0, len(info))
	for _, node := range info {
		nodes = append(nodes, ToBasicNodeInfo(node))
	}

	return nodes
}

func ToBasicNodeInfo(info domain.NodeBaseInfo) *pb.NodeBaseInfo {
	return &pb.NodeBaseInfo{
		Id:     info.Id,
		Name:   info.Name,
		Addr:   info.Addr,
		Status: NodeStatusFromLastOnline(info),
	}
}

func NodeStatusFromLastOnline(inf domain.NodeBaseInfo) pb.NodeStatus {
	if !inf.IsEnabled {
		return pb.NodeStatus_NodeStatus_Offline
	}

	if inf.LastOnline.Before(time.Now().Add(-time.Minute)) {
		return pb.NodeStatus_NodeStatus_Degraded
	}

	return pb.NodeStatus_NodeStatus_Online
}
