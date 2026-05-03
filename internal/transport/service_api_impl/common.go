package service_api_impl

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func toServiceBaseInfoList(in []domain.ServiceBaseInfo) []*pb.ServiceBaseInfo {
	out := make([]*pb.ServiceBaseInfo, 0, len(in))
	for _, v := range in {
		out = append(out, toServiceBaseInfo(v))
	}

	return out
}

func toServiceBaseInfo(in domain.ServiceBaseInfo) *pb.ServiceBaseInfo {
	info := &pb.ServiceBaseInfo{
		Name: in.Name,
	}

	if in.LastDeployedAt != nil {
		info.LastDeployedAt = timestamppb.New(*in.LastDeployedAt)
	}

	return info
}
