package transfer_out

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func beginRequest(begin *velez_api.Transfer_Begin) *velez_api.Transfer_Request {
	frame := &velez_api.Transfer_Request_Begin{Begin: begin}

	return &velez_api.Transfer_Request{Frame: frame}
}

func itemStartRequest(name string) *velez_api.Transfer_Request {
	start := &velez_api.Transfer_ItemStart{Name: name}
	frame := &velez_api.Transfer_Request_ItemStart{ItemStart: start}

	return &velez_api.Transfer_Request{Frame: frame}
}

func batchRequest(data []byte) *velez_api.Transfer_Request {
	batch := &velez_api.Transfer_Batch{Data: data}
	frame := &velez_api.Transfer_Request_Batch{Batch: batch}

	return &velez_api.Transfer_Request{Frame: frame}
}

func endRequest() *velez_api.Transfer_Request {
	frame := &velez_api.Transfer_Request_End{End: &velez_api.Transfer_End{}}

	return &velez_api.Transfer_Request{Frame: frame}
}
