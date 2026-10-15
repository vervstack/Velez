package transfer_in

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func newAcceptedResponse() *velez_api.Transfer_Response {
	return &velez_api.Transfer_Response{
		Frame: &velez_api.Transfer_Response_Accepted{Accepted: &velez_api.Transfer_Accepted{}},
	}
}

func newItemAppliedResponse(name string) *velez_api.Transfer_Response {
	return &velez_api.Transfer_Response{
		Frame: &velez_api.Transfer_Response_ItemApplied{
			ItemApplied: &velez_api.Transfer_ItemApplied{Name: name},
		},
	}
}

func newFinishedResponse() *velez_api.Transfer_Response {
	return &velez_api.Transfer_Response{
		Frame: &velez_api.Transfer_Response_Finished{Finished: &velez_api.Transfer_Finished{}},
	}
}
