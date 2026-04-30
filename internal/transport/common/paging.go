package common

import (
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func FromPaging(in *pb.Paging) domain.Paging {
	return domain.Paging{
		Limit:  in.GetLimit(),
		Offset: in.GetOffset(),
	}
}
