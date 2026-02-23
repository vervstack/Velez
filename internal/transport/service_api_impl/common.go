package service_api_impl

import (
	"go.vervstack.ru/Velez/internal/domain"
	pb "go.vervstack.ru/Velez/pkg/velez_api"
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
