package common

import (
	"go.vervstack.ru/Velez/internal/domain"
	pb "go.vervstack.ru/Velez/pkg/velez_api"
)

func FromPaging(in *pb.Paging) domain.Paging {
	return domain.Paging{
		Limit:  in.GetLimit(),
		Offset: in.GetOffset(),
	}
}
