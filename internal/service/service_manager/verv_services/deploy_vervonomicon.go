package verv_services

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// CreateDeployFromVervonomicon reads the .verv/ descriptor baked into
// req.Image, merges the req.Environment overlay, resolves it into a
// CreateSmerd.Request and proceeds exactly as CreateNewDeploy does from
// there - per docs/features/vervonomicon.md, only the primary app container
// is resolved this wave; ingress is a later wave's concern.
//
// resources[] is reconciled (see reconcileResources) but not yet provisioned:
// no resource type in this codebase has a reusable per-service provisioning
// path today (see docs/features/vervonomicon.md's "Resource reconciliation"
// and the wave 3b report) - an already_connected resource correctly gets
// nothing created, and a must_provision one is logged and left for a later
// wave to actually provision, rather than inventing a new provisioner here.
func (v *VervService) CreateDeployFromVervonomicon(
	ctx context.Context, req domain.CreateDeployFromVervonomiconReq,
) error {
	descriptor, _, err := v.readVervonomicon(ctx, req.ServiceName, req.Image, req.Environment)
	if err != nil {
		if errors.Is(err, vervonomicon.ErrNoDescriptor) {
			return rerrors.Wrap(user_errors.ErrNoVervonomiconDescriptor)
		}

		return rerrors.Wrap(err, "error reading vervonomicon descriptor")
	}

	resourceDecisions, err := v.reconcileResources(ctx, req.ServiceName, descriptor.Resources)
	if err != nil {
		return rerrors.Wrap(err, "error reconciling vervonomicon resources")
	}

	logResourceDecisions(req.ServiceName, resourceDecisions)

	resolver := vervonomicon.NewBoxResolver(v.boxes())

	smerdRequest, err := resolver.ResolveRequest(ctx, descriptor, req.Environment, req.Image)
	if err != nil {
		return rerrors.Wrap(err, "error resolving vervonomicon deployment request")
	}

	deployReq := domain.CreateDeployReq{
		ServiceName:    req.ServiceName,
		VervDescriptor: &descriptor,
		LaunchSmerd: domain.LaunchSmerd{
			CreateSmerd_Request: smerdRequest,
		},
	}

	err = v.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating vervonomicon deploy")
	}

	return nil
}

// logResourceDecisions surfaces every resources.yaml entry's reconciliation
// outcome at deploy time. A must_provision entry is not acted on this wave -
// see CreateDeployFromVervonomicon's doc comment - so it's logged as a
// warning to make the deferral visible rather than silent.
func logResourceDecisions(serviceName string, decisions []domain.ResourceReconciliation) {
	for _, decision := range decisions {
		if decision.Status == domain.ResourceMustProvision {
			logResourceDecision(log.Warn(), serviceName, decision)

			continue
		}

		logResourceDecision(log.Info(), serviceName, decision)
	}
}

func logResourceDecision(event *zerolog.Event, serviceName string, decision domain.ResourceReconciliation) {
	event.
		Str("service_name", serviceName).
		Str("resource_name", decision.Name).
		Str("resource_type", decision.Type).
		Str("status", string(decision.Status)).
		Msg("vervonomicon resource reconciliation decision")
}
