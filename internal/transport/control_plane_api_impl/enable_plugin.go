package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/transport/common"
)

var (
	//nolint:forbidigo // package-private sentinel, not shared/user-facing
	errUnsupportedService = rerrors.New("unsupported service", codes.InvalidArgument)

	//nolint:forbidigo // package-private sentinel, not shared/user-facing
	errInvalidPayload = rerrors.New("invalid payload", codes.InvalidArgument)
)

func (impl *Impl) EnablePlugin(ctx context.Context, req *pb.EnablePlugin_Request) (
	*pb.EnablePlugin_Response, error,
) {
	switch req.GetPlugin() {
	case pb.VervPluginType_statefull_pg:
		payload, ok := req.GetPayload().(*pb.EnablePlugin_Request_StatefullCluster)
		if !ok {
			return nil, rerrors.Wrap(errInvalidPayload)
		}

		// state.PgName("") is this jobs-engine task's entityID (a key into
		// this node's own local task storage, not a Docker resource name) -
		// intentionally never suffixed, since a node only ever calls
		// EnablePlugin for its own cluster once. There's exactly one
		// statefull_pg task per node, so it doubles as the task's entityID
		// (mirrors CreateSmerd/AssembleConfig/ConnectService keying tasks off
		// their own natural per-request identifier).
		initialContext := &pb.EnableStatefullTaskPayload{
			Request: payload.StatefullCluster,
		}

		entityID := state.PgName("")

		_, err := impl.jobsEngine.Enqueue(ctx, entityID, jobs.EnableStatefullAction, initialContext)
		if err != nil {
			return &pb.EnablePlugin_Response{}, rerrors.Wrap(err, "error during enabling plugin")
		}

		resp := &pb.EnablePlugin_Response{
			EntityId: entityID,
			Action:   jobs.EnableStatefullAction,
		}

		return resp, nil
	case pb.VervPluginType_headscale:
		return impl.enableHeadscale(ctx, req.GetHeadscaleServer())
	default:
		return nil, rerrors.Wrap(errUnsupportedService)
	}
}

func (impl *Impl) enableHeadscale(ctx context.Context, request *pb.EnableHeadscaleServer) (
	*pb.EnablePlugin_Response, error,
) {
	input := &pb.CreateVcnNetwork_Request{Name: "headscale"}

	switch {
	case request.GetDeployConfig() != nil:
		config := request.GetDeployConfig()
		deploy := &pb.CreateVcnNetwork_Deploy{CustomPort: config.CustomPort, CustomImage: config.CustomImage}

		input.Source = &pb.CreateVcnNetwork_Request_Deploy{Deploy: deploy}
	case request.GetExternalConnect() != nil:
		config := request.GetExternalConnect()
		external := &pb.CreateVcnNetwork_External{Url: config.GetUrl(), Token: config.GetToken()}

		input.Source = &pb.CreateVcnNetwork_Request_External{External: external}
	default:
		return nil, rerrors.Wrap(errInvalidPayload)
	}

	domainReq := common.CreateVcnNetworkRequest(ctx, input)

	network, err := impl.networks.Create(ctx, domainReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &pb.EnablePlugin_Response{NetworkId: &network.Id}, nil
}
