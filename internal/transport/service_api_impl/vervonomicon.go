package service_api_impl

import (
	"context"
	"sort"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

// GetVervonomicon decodes the request, calls the service layer, and encodes
// the result - all resolution logic (finding the current image, reading and
// merging the descriptor, masking sensitive values) lives in
// service_manager/verv_services.VervService.GetVervonomicon.
//
// A NoDescriptor result is encoded as a plain empty response, never a gRPC
// error, per docs/features/vervonomicon.md: "Absence of a descriptor is
// never an error."
func (impl *Impl) GetVervonomicon(
	ctx context.Context,
	pbReq *pb.GetVervonomicon_Request,
) (*pb.GetVervonomicon_Response, error) {
	req := domain.GetVervonomiconReq{
		ServiceName: pbReq.GetServiceName(),
		Environment: pbReq.GetEnvironment(),
	}

	result, err := impl.servicesService.GetVervonomicon(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting vervonomicon")
	}

	if result.NoDescriptor {
		return &pb.GetVervonomicon_Response{}, nil
	}

	resp := &pb.GetVervonomicon_Response{
		Raw:              toDescriptorFiles(result.Raw),
		ResolvedYaml:     result.ResolvedYaml,
		Source:           toVervonomiconSource(result.Source),
		Environment:      result.Environment,
		ResourceStatuses: toResourceReconciliations(result.ResourceStatuses),
	}

	return resp, nil
}

// toDescriptorFiles sorts by path for a deterministic response - the source
// map has no inherent order.
func toDescriptorFiles(files map[string][]byte) []*pb.DescriptorFile {
	if len(files) == 0 {
		return nil
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}

	sort.Strings(paths)

	out := make([]*pb.DescriptorFile, len(paths))
	for i, path := range paths {
		out[i] = &pb.DescriptorFile{
			Path:    path,
			Content: files[path],
		}
	}

	return out
}

func toVervonomiconSource(source verv.SourceKind) pb.VervonomiconSource {
	switch source {
	case verv.SourceKindImage:
		return pb.VervonomiconSource_VERVONOMICON_SOURCE_IMAGE
	case verv.SourceKindRepo:
		return pb.VervonomiconSource_VERVONOMICON_SOURCE_REPO
	case verv.SourceKindPushed:
		return pb.VervonomiconSource_VERVONOMICON_SOURCE_PUSHED
	default:
		return pb.VervonomiconSource_VERVONOMICON_SOURCE_UNSPECIFIED
	}
}

func toResourceReconciliations(decisions []domain.ResourceReconciliation) []*pb.ResourceReconciliation {
	if len(decisions) == 0 {
		return nil
	}

	out := make([]*pb.ResourceReconciliation, len(decisions))
	for i, decision := range decisions {
		out[i] = &pb.ResourceReconciliation{
			Name:         decision.Name,
			ResourceType: decision.Type,
			Status:       toResourceConnectionStatus(decision.Status),
		}
	}

	return out
}

func toResourceConnectionStatus(status domain.ResourceConnectionStatus) pb.ResourceConnectionStatus {
	switch status {
	case domain.ResourceAlreadyConnected:
		return pb.ResourceConnectionStatus_RESOURCE_CONNECTION_STATUS_ALREADY_CONNECTED
	case domain.ResourceMustProvision:
		return pb.ResourceConnectionStatus_RESOURCE_CONNECTION_STATUS_MUST_PROVISION
	default:
		return pb.ResourceConnectionStatus_RESOURCE_CONNECTION_STATUS_UNSPECIFIED
	}
}
