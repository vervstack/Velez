package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
)

// pgCredentials is the generated db_name/username/password triple for one PG
// instance. Only buildDeployRequest ever sees the password - it lands on the
// resolved CreateSmerd.Request's env, never on the descriptor.
type pgCredentials struct {
	dbName   string
	username string
	password string
}

// buildDeployRequest reads the builtin postgres descriptor, overlays the
// instance's shape (box, unique volume name, exposed port), resolves it into
// a CreateSmerd.Request via the box resolver, then overlays the instance's
// generated name and credentials onto the *resolved request* - never onto
// the descriptor. Per docs/features/pgaas_and_registry_plugin.md section 1's
// "no descriptor file ever contains a credential" rule and section 3's
// create flow: "The descriptor deliberately carries no credentials".
//
// The returned descriptor is what CreateDeployReq.VervDescriptor persists;
// the returned request is what actually launches the container.
func buildDeployRequest(
	ctx context.Context,
	boxes vervonomicon.BoxLookup,
	req domain.CreatePgInstanceReq,
	creds pgCredentials,
) (verv.Descriptor, *velez_api.CreateSmerd_Request, error) {
	files, err := builtin.Read("postgres")
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error reading builtin postgres descriptor")
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, req.Environment)
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error merging builtin postgres descriptor environment")
	}

	descriptor.Source = verv.SourceKindBuiltin

	if req.Box != "" {
		descriptor.Deployment.App.Box = req.Box
	}

	if len(descriptor.Deployment.App.Volumes) > 0 {
		descriptor.Deployment.App.Volumes[0].Name = pgVolumeName(req.Name)
	}

	if req.ExposeToPort != 0 && len(descriptor.Deployment.App.Ports) > 0 {
		descriptor.Deployment.App.Ports[0].ExposeTo = int(req.ExposeToPort)
	}

	resolver := vervonomicon.NewBoxResolver(boxes)

	request, err := resolver.ResolveRequest(ctx, descriptor, req.Environment, "")
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error resolving pg instance deploy request")
	}

	request.Name = req.Name
	request.Env = map[string]string{
		"POSTGRES_DB":       creds.dbName,
		"POSTGRES_USER":     creds.username,
		"POSTGRES_PASSWORD": creds.password,
	}

	if request.Labels == nil {
		request.Labels = make(map[string]string)
	}

	// VervServiceLabel makes the instance a first-class entry in the node's
	// service list (listDistinctServices keys on it); PgaasInstanceLabel lets
	// the single-node local_storage backend recover the instance's facts from
	// the running container - see internal/storage/local_storage/pg_instances.go.
	// Both are inert in cluster mode, where velez.pg_instances is authoritative.
	request.Labels[labels.VervServiceLabel] = req.Name
	request.Labels[labels.PgaasInstanceLabel] = "true"

	return descriptor, request, nil
}
