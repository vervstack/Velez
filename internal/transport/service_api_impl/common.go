package service_api_impl

import (
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
	return &pb.ServiceBaseInfo{
		Name: in.Name,
	}
}
