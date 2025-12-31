package cluster

import (
	"context"

	"go.vervstack.ru/makosh/pkg/makosh_be"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
)

type disabledConfigurator struct {
}

func (d disabledConfigurator) ApiVersion(_ context.Context, _ *matreshka_api.ApiVersion_Request, _ ...grpc.CallOption) (*matreshka_api.ApiVersion_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) GetConfig(_ context.Context, _ *matreshka_api.GetConfig_Request, _ ...grpc.CallOption) (*matreshka_api.GetConfig_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) GetConfigNodes(_ context.Context, _ *matreshka_api.GetConfigNode_Request, _ ...grpc.CallOption) (*matreshka_api.GetConfigNode_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) ListConfigs(_ context.Context, _ *matreshka_api.ListConfigs_Request, _ ...grpc.CallOption) (*matreshka_api.ListConfigs_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) CreateConfig(_ context.Context, _ *matreshka_api.CreateConfig_Request, _ ...grpc.CallOption) (*matreshka_api.CreateConfig_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) PatchConfig(_ context.Context, _ *matreshka_api.PatchConfig_Request, _ ...grpc.CallOption) (*matreshka_api.PatchConfig_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) StoreConfig(_ context.Context, _ *matreshka_api.StoreConfig_Request, _ ...grpc.CallOption) (*matreshka_api.StoreConfig_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) RenameConfig(_ context.Context, _ *matreshka_api.RenameConfig_Request, _ ...grpc.CallOption) (*matreshka_api.RenameConfig_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) DeleteConfig(_ context.Context, _ *matreshka_api.DeleteConfig_Request, _ ...grpc.CallOption) (*matreshka_api.DeleteConfig_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledConfigurator) SubscribeOnChanges(_ context.Context, _ ...grpc.CallOption) (grpc.BidiStreamingClient[matreshka_api.SubscribeOnChanges_Request, matreshka_api.SubscribeOnChanges_Response], error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

type disabledServiceDiscovery struct {
}

func (d disabledServiceDiscovery) Version(_ context.Context, _ *makosh_be.Version_Request, _ ...grpc.CallOption) (*makosh_be.Version_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledServiceDiscovery) ListEndpoints(_ context.Context, _ *makosh_be.ListEndpoints_Request, _ ...grpc.CallOption) (*makosh_be.ListEndpoints_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (d disabledServiceDiscovery) UpsertEndpoints(_ context.Context, _ *makosh_be.UpsertEndpoints_Request, _ ...grpc.CallOption) (*makosh_be.UpsertEndpoints_Response, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}
