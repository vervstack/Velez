package runneraas

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
)

const (
	runnerSecretScope = "runneraas"
	runnerSecretKey   = "access_token"
)

func (s *RunneraasService) CreateRunner(ctx context.Context, req domain.CreateRunnerReq) (domain.RunnerView, error) {
	err := validateRunnerTarget(req.Scope, req.Target)
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err)
	}

	provider, err := s.provider(req.Provider)
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err)
	}

	secretRef := domain.SecretRef{Scope: runnerSecretScope, Owner: req.Name, Key: runnerSecretKey}

	err = s.secrets.Put(ctx, secretRef, req.AccessToken)
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err, "error storing runner access token")
	}

	token, err := provider.MintRegistrationToken(ctx, req.Scope, req.Target, req.AccessToken)
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err, "error minting runner registration token")
	}

	descriptor, smerdRequest, err := s.buildDeployRequest(ctx, provider, req, token)
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err, "error building runner deploy request")
	}

	// Docker-socket grant MUST be written before CreateNewDeploy schedules the
	// deploy - deploy_watcher.go reads it, by the same DockerSocketGrantSecretRef
	// derivation, right before it enqueues the create_smerd task that actually
	// creates the container. This is the only writer of this secret in the
	// codebase (besides jobs/enable_github_runner.go's now-superseded path) -
	// see create_smerd.go's dockerSocketAccessor gate. Never derived from, or
	// settable via, any field on CreateRunner.Request.
	err = s.secrets.Put(ctx, domain.DockerSocketGrantSecretRef(req.Name), "true")
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err, "error putting docker socket grant secret")
	}

	deployReq := domain.CreateDeployReq{
		ServiceName:    req.Name,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	// CreateNewDeploy upserts req.Name before looking it up, so no separate
	// UpsertService call is needed here (see verv_services/deploy.go).
	err = s.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err, "error creating runner deploy")
	}

	svc, err := s.dataStorage.Services().GetByName(ctx, req.Name)
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err, "error getting runner service")
	}

	upsertReq := domain.UpsertRunnerReq{
		ServiceID: svc.ID,
		Provider:  req.Provider.String(),
		Scope:     req.Scope.String(),
		Target:    req.Target,
		Labels:    req.Labels,
		SecretRef: secretRef.String(),
	}

	instance, err := s.dataStorage.Runners().UpsertRunner(ctx, upsertReq)
	if err != nil {
		return domain.RunnerView{}, rerrors.Wrap(err, "error upserting runner row")
	}

	view := domain.RunnerView{
		Name:        req.Name,
		Provider:    req.Provider,
		Scope:       req.Scope,
		Target:      instance.Target,
		Labels:      instance.Labels,
		Environment: req.Environment,
		CreatedAt:   instance.CreatedAt,
		UpdatedAt:   instance.UpdatedAt,
	}

	return view, nil
}

// buildDeployRequest reads the builtin descriptor named by provider, overlays
// the instance's shape (unique volume name, registration env), resolves it
// into a CreateSmerd.Request via the box resolver, then overlays the
// instance's generated name onto the *resolved request* - never onto the
// descriptor. Mirrors pgaas.buildDeployRequest's "no descriptor file ever
// contains a credential" discipline.
//
// The returned descriptor is what CreateDeployReq.VervDescriptor persists;
// the returned request is what actually launches the container.
func (s *RunneraasService) buildDeployRequest(
	ctx context.Context, provider Provider, req domain.CreateRunnerReq, registrationToken string,
) (verv.Descriptor, *velez_api.CreateSmerd_Request, error) {
	files, err := builtin.Read(provider.DescriptorName())
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error reading builtin runner descriptor")
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, req.Environment)
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error merging builtin runner descriptor environment")
	}

	descriptor.Source = verv.SourceKindBuiltin

	for i := range descriptor.Deployment.App.Volumes {
		if descriptor.Deployment.App.Volumes[i].Path != provider.DataPath() {
			continue
		}

		descriptor.Deployment.App.Volumes[i].Name = runnerVolumeName(req.Name)
	}

	resolver := vervonomicon.NewBoxResolver(s.boxes())

	request, err := resolver.ResolveRequest(ctx, descriptor, req.Environment, "")
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error resolving runner deploy request")
	}

	request.Name = req.Name

	runnerName := deriveRunnerName(req.Target)

	request.Env = provider.RegistrationEnv(req.Scope, req.Target, runnerName, registrationToken, req.Labels)

	if request.Labels == nil {
		request.Labels = make(map[string]string)
	}

	// VervServiceLabel makes the instance a first-class entry in the node's
	// service list (listDistinctServices keys on it); the Runner*Label set
	// lets the single-node local_storage backend recover the instance's
	// facts from the running container directly, without parsing any
	// provider-specific env var - see
	// internal/storage/local_storage/runners.go and
	// internal/domain/labels.RunnerInstanceLabel's doc comment. All inert in
	// cluster mode, where velez.runners is authoritative.
	request.Labels[labels.VervServiceLabel] = req.Name
	request.Labels[labels.RunnerInstanceLabel] = "true"
	request.Labels[labels.RunnerProviderLabel] = req.Provider.String()
	request.Labels[labels.RunnerScopeLabel] = req.Scope.String()
	request.Labels[labels.RunnerTargetLabel] = req.Target
	request.Labels[labels.RunnerLabelsLabel] = strings.Join(req.Labels, ",")

	return descriptor, request, nil
}
