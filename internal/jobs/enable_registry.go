package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"path"
	"strconv"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"golang.org/x/crypto/bcrypt"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
	"go.vervstack.ru/Velez/internal/storage/registries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	EnableRegistryAction = "enable_registry"

	// RegistryServiceName is the fixed name of the single, node-wide
	// registry plugin's service/container - unlike a PG instance (section 3
	// of docs/features/pgaas_and_registry_plugin.md), a node only ever runs
	// one registry, so there's no per-instance naming to resolve.
	RegistryServiceName = "registry"

	registryDefaultUsername = "verv"

	// registryDescriptorName selects the builtin descriptor - see
	// internal/service/service_manager/vervonomicon/builtin.
	registryDescriptorName = "registry"

	// registryAuthVolumeName/registryHtpasswdPath name the named volume and
	// in-container path the registry plugin's htpasswd file is written to -
	// must match builtin/registry/deployment.yaml's "registry-auth" volume
	// (mounted at /auth).
	registryAuthVolumeName = "registry-auth"
	registryHtpasswdPath   = "/auth/htpasswd"

	// registryContainerPort is the registry image's own listening port
	// (builtin/registry/deployment.yaml's app.ports[0].port) - never the
	// host-exposed port, which is resolved per enable request.
	registryContainerPort = 5000

	// registryEnvironment mirrors enable_statefull.go's statefullEnvironment:
	// EnablePlugin carries no `environment` field (it's a node-level
	// operation), so the registry always deploys into the default
	// environment.
	registryEnvironment = environments.DefaultEnvironmentName

	// envRegistryAuth* overlay the registry:2 image's htpasswd basic-auth env
	// vars - builtin/registry/deployment.yaml deliberately carries no env
	// (per docs/features/vervonomicon.md, no descriptor file ever contains a
	// credential), so this job chain overlays them at deploy time instead.
	envRegistryAuth              = "REGISTRY_AUTH"
	envRegistryAuthHtpasswdRealm = "REGISTRY_AUTH_HTPASSWD_REALM"
	envRegistryAuthHtpasswdPath  = "REGISTRY_AUTH_HTPASSWD_PATH"
	registryAuthRealm            = "Registry Realm"

	secretScopePlugin   = "plugin"
	secretOwnerRegistry = "registry"
	secretKeyPassword   = "password"

	stepPutRegistrySecret = "put_secret"
	stepWriteHtpasswd     = "write_htpasswd"
	stepDeployRegistry    = "deploy_registry"
	stepRegisterPlugin    = "register_plugin"
	stepRegisterRegistry  = "register_registry"
)

// Accessor interfaces the enable_registry jobs need from their TaskContext.
// *velez_api.EnableRegistryTaskPayload satisfies all of them.
// containerIDAccessor is declared in create_smerd.go and reused here, exactly
// as copy_to_volume.go and enable_statefull.go already do.

type registryRequestAccessor interface {
	GetRequest() *velez_api.EnableRegistry
}

type registryUsernameAccessor interface {
	GetUsername() string
	SetUsername(username string)
}

type registryPasswordAccessor interface {
	GetPassword() string
	SetPassword(password string)
}

type registryPortAccessor interface {
	GetExposedPort() uint32
	SetExposedPort(port uint32)
}

// registryPortManager is the narrow node_clients.PortManager slice
// deployRegistryJob needs: auto-assigning a free port from the node's shared
// pool when the caller doesn't request an explicit expose_to_port, then
// marking it on hold (HoldPort) so create_smerd's own lockPorts step - which
// re-locks every port carrying a non-nil ExposedTo - finds it via UnHoldPort
// and skips re-locking it instead of colliding with the lock this job already
// took. Declared narrow, following startAPI/copyAPI's precedent in
// copy_to_volume.go, so this job stays hand-fakeable without depending on the
// full PortManager interface.
type registryPortManager interface {
	GetPortForEnvironment(environment string) (uint32, error)
	HoldPort(port uint32) bool
}

type enableRegistryHandler struct {
	nodeClients node_clients.NodeClients
	runtimes    container_runtime.RuntimeResolver
	// storageContainer must be the same storage.Storage instance
	// VervServicesService.CreateNewDeploy reads/writes (cluster_clients'
	// StateManager, not service_manager's own StorageContainer) - the two
	// only converge onto the same Postgres-backed storage once this node has
	// joined a Postgres cluster (see custom.go's InitServiceLayer and
	// enable_statefull.go's updateClusterStateJob), and deployRegistryJob
	// hands off to CreateNewDeploy, so every other storage read/write in this
	// job chain has to agree with it from the start, not just after
	// convergence.
	storageContainer storage.Storage
	secretsStore     secrets.Store
	vervServices     service.VervServicesService
}

func NewEnableRegistryHandler(
	nodeClients node_clients.NodeClients,
	runtimes container_runtime.RuntimeResolver,
	storageContainer storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
) TaskHandler {
	return &enableRegistryHandler{
		nodeClients:      nodeClients,
		runtimes:         runtimes,
		storageContainer: storageContainer,
		secretsStore:     secretsStore,
		vervServices:     vervServices,
	}
}

func (h *enableRegistryHandler) Action() string {
	return EnableRegistryAction
}

func (h *enableRegistryHandler) NewContext() TaskContext {
	return &velez_api.EnableRegistryTaskPayload{}
}

// BuildJobs, per docs/features/pgaas_and_registry_plugin.md section 4:
// generate_credentials/put_secret persist a username+password the same way
// enable_statefull.go's generateCredentialsJob does; the four
// create/start/write/drop steps write a bcrypt htpasswd entry into the
// registry-auth volume by reusing copy_to_volume.go's loader-container jobs
// verbatim (createLoaderContainerJob/startLoaderContainerJob/
// dropLoaderContainerJob and its writeFileToContainer helper) - only the
// per-file byte-content step is new (writeHtpasswdJob), because
// copyFileJob's content is a plain []byte snapshotted at BuildJobs time,
// before generate_credentials has run, and so cannot carry a value generated
// later in the same task attempt; deploy_registry resolves the builtin
// descriptor and hands off to the existing CreateNewDeploy path - the deploy
// watcher creates the actual container, this handler never does; and
// register_plugin/register_registry record the plugin the same way
// enable_statefull.go's registerPluginJob does, plus the registries row that
// makes the registry immediately usable for image pulls/pushes.
func (h *enableRegistryHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.EnableRegistryTaskPayload)
	if !ok {
		panic("enable_registry: BuildJobs called with mismatched TaskContext type")
	}

	folders := mountedFolders(registryAuthVolumeName, []string{registryHtpasswdPath})

	return []NamedJob{
		{
			Name: stepGenerateCredentials,
			Job: &generateRegistryCredentialsJob{
				req: payload,
				ctx: payload,
			},
		},
		{
			Name: stepPutRegistrySecret,
			Job: &putRegistrySecretJob{
				secrets: h.secretsStore,
				ctx:     payload,
			},
		},
		{
			Name: stepCreateLoaderContainer,
			Job: &createLoaderContainerJob{
				nodeClients: h.nodeClients,
				req:         registryVolumeRef{},
				folders:     folders,
				ctx:         payload,
			},
		},
		{
			Name: stepStartSidecar,
			Job: &startLoaderContainerJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				ctx:       payload,
			},
		},
		{
			Name: stepWriteHtpasswd,
			Job: &writeHtpasswdJob{
				runtimes: h.runtimes,
				copyAPI:  h.nodeClients.Docker().Client(),
				ctx:      payload,
				filePath: registryHtpasswdPath,
			},
		},
		{
			Name: stepDropContainer,
			Job: &dropLoaderContainerJob{
				docker: h.nodeClients.Docker(),
				ctx:    payload,
			},
		},
		{
			Name: stepDeployRegistry,
			Job: &deployRegistryJob{
				boxes:        h.storageContainer.ResourceBoxes(),
				portManager:  h.nodeClients.PortManager(),
				vervServices: h.vervServices,
				req:          payload,
				ctx:          payload,
			},
		},
		{
			Name: stepRegisterPlugin,
			Job: &registerRegistryPluginJob{
				storageContainer: h.storageContainer,
			},
		},
		{
			Name: stepRegisterRegistry,
			Job: &registerRegistryRowJob{
				registries: h.storageContainer.Registries(),
				ctx:        payload,
			},
		},
	}
}

// registryVolumeRef adapts the fixed registry-auth volume name to
// copyToVolumeRequestAccessor (declared in copy_to_volume.go) so
// createLoaderContainerJob can be reused verbatim for the htpasswd loader
// container - GetPathToFiles is never called by createLoaderContainerJob.Do,
// only GetVolumeName is, so a nil map here is fine.
type registryVolumeRef struct{}

func (registryVolumeRef) GetVolumeName() string {
	return registryAuthVolumeName
}

func (registryVolumeRef) GetPathToFiles() map[string][]byte {
	return nil
}

// generateRegistryCredentialsJob mirrors enable_statefull.go's
// generateCredentialsJob: derives the registry's basic-auth username/password
// once and persists them into the task's context, so a crash-resumed task
// reuses the same credentials instead of regenerating ones that would
// mismatch the htpasswd entry already written into the registry-auth volume.
type generateRegistryCredentialsJob struct {
	req registryRequestAccessor

	ctx interface {
		registryUsernameAccessor
		registryPasswordAccessor
	}
}

func (j *generateRegistryCredentialsJob) Do(_ context.Context) error {
	if j.ctx.GetUsername() == "" {
		username := j.req.GetRequest().GetUsername()
		if username == "" {
			username = registryDefaultUsername
		}

		j.ctx.SetUsername(username)
	}

	if j.ctx.GetPassword() == "" {
		j.ctx.SetPassword(string(toolbox.RandomBase64(generatedPwdLength)))
	}

	return nil
}

// putRegistrySecretJob persists the generated password into the secret store
// under the plugin/registry/password ref - see
// docs/features/pgaas_and_registry_plugin.md section 1. No Rollback: the
// secret is keyed by a fixed ref, so a later Put (from a resumed task, or a
// future re-enable) simply overwrites it rather than needing to be undone.
type putRegistrySecretJob struct {
	secrets secrets.Store

	ctx registryPasswordAccessor
}

func (j *putRegistrySecretJob) Do(ctx context.Context) error {
	ref := registryPasswordSecretRef()

	err := j.secrets.Put(ctx, ref, j.ctx.GetPassword())
	if err != nil {
		return rerrors.Wrap(err, "error putting registry password secret")
	}

	return nil
}

func registryPasswordSecretRef() domain.SecretRef {
	return domain.SecretRef{
		Scope: secretScopePlugin,
		Owner: secretOwnerRegistry,
		Key:   secretKeyPassword,
	}
}

// writeHtpasswdJob writes the bcrypt htpasswd entry into the loader
// container created by createLoaderContainerJob, reusing
// copy_to_volume.go's writeFileToContainer tar-write helper verbatim (same
// mkdir-then-write shape as copyFileJob.Do). It can't reuse copyFileJob
// itself: copyFileJob.content is a plain []byte fixed at BuildJobs time,
// before generate_credentials has run, so it can never carry a
// generated-this-run password - see BuildJobs's doc comment.
type writeHtpasswdJob struct {
	runtimes container_runtime.RuntimeResolver
	copyAPI  copyAPI

	ctx interface {
		containerIDAccessor
		registryUsernameAccessor
		registryPasswordAccessor
	}

	filePath string
}

func (j *writeHtpasswdJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return user_errors.ErrContainerIdMissing
	}

	line, err := htpasswdLine(j.ctx.GetUsername(), j.ctx.GetPassword())
	if err != nil {
		return rerrors.Wrap(err, "error generating htpasswd entry")
	}

	containerRuntime, err := j.runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = mkdirInContainer(ctx, containerRuntime, containerID, path.Dir(j.filePath))
	if err != nil {
		return err
	}

	err = writeFileToContainer(ctx, j.copyAPI, containerID, j.filePath, []byte(line))
	if err != nil {
		return rerrors.Wrap(err, "error copying htpasswd to container")
	}

	return nil
}

// htpasswdLine renders one apache-htpasswd-format line ("user:bcryptHash\n"),
// the format REGISTRY_AUTH=htpasswd expects - bcrypt is the only hash the
// registry:2 image's htpasswd auth driver accepts.
func htpasswdLine(username, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", rerrors.Wrap(err, "error hashing registry password")
	}

	return fmt.Sprintf("%s:%s\n", username, string(hash)), nil
}

// deployRegistryJob resolves the builtin registry descriptor, overlays the
// generated deploy's exposed port and auth env vars, and hands off to the
// same VervServicesService.CreateNewDeploy path any other vervonomicon
// service deploys through (see
// internal/service/service_manager/verv_services/deploy_vervonomicon.go) -
// the deploy watcher creates and starts the actual container from there.
// This job never creates a container itself.
type deployRegistryJob struct {
	boxes        vervonomicon.BoxLookup
	portManager  registryPortManager
	vervServices service.VervServicesService

	req registryRequestAccessor
	ctx registryPortAccessor
}

func (j *deployRegistryJob) Do(ctx context.Context) error {
	files, err := builtin.Read(registryDescriptorName)
	if err != nil {
		return rerrors.Wrap(err, "error reading builtin registry descriptor")
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, registryEnvironment)
	if err != nil {
		return rerrors.Wrap(err, "error merging registry vervonomicon environment overlay")
	}

	descriptor.Source = verv.SourceKindBuiltin

	resolver := vervonomicon.NewBoxResolver(j.boxes)

	smerdRequest, err := resolver.ResolveRequest(ctx, descriptor, registryEnvironment, "")
	if err != nil {
		return rerrors.Wrap(err, "error resolving registry vervonomicon deployment request")
	}

	port, err := j.resolvePort()
	if err != nil {
		return err
	}

	overlayRegistryPort(smerdRequest, port)
	overlayRegistryAuthEnv(smerdRequest)

	j.ctx.SetExposedPort(port)

	deployReq := domain.CreateDeployReq{
		ServiceName:    RegistryServiceName,
		VervDescriptor: &descriptor,
		LaunchSmerd: domain.LaunchSmerd{
			CreateSmerd_Request: smerdRequest,
		},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating registry deploy")
	}

	return nil
}

// resolvePort returns the caller's requested expose_to_port verbatim -
// create_smerd's own lockPorts job locks it for real once the deploy watcher
// actually runs create_smerd (see BuildJobs's doc comment on this handler
// never creating a container itself) - or, when none was requested,
// auto-assigns and locks one from the node's shared port pool right now,
// since deploy_registry needs a concrete port immediately to build the
// registries.url row. That lock is put on hold (HoldPort) rather than left
// bare: create_smerd's lockPorts job runs against the same port later (it
// carries a non-nil ExposedTo via overlayRegistryPort) and, seeing an
// already-locked port with no hold recorded, would call
// LockPortForEnvironment against a port it doesn't own yet and fail with
// ErrPortAlreadyLocked - LockPortForEnvironment has no "already locked by
// this same environment" no-op path, it always rejects a locked port. Holding
// it here makes lockPorts's UnHoldPort check consume the hold and skip
// re-locking instead.
func (j *deployRegistryJob) resolvePort() (uint32, error) {
	requested := j.req.GetRequest().GetExposeToPort()
	if requested != 0 {
		return requested, nil
	}

	port, err := j.portManager.GetPortForEnvironment(registryEnvironment)
	if err != nil {
		return 0, rerrors.Wrap(err, "error resolving free port for registry")
	}

	j.portManager.HoldPort(port)

	return port, nil
}

// overlayRegistryPort sets ExposedTo on the descriptor-resolved port entry
// matching registryContainerPort. Without this, the port is never published
// to the host at all (see internal/clients/node_clients/docker/dockerutils/
// parser.FromPorts: ExposedTo == nil is silently skipped, "TODO auto assign
// if not exists" - not implemented there).
func overlayRegistryPort(request *velez_api.CreateSmerd_Request, port uint32) {
	for _, p := range request.GetSettings().GetPorts() {
		if p.GetServicePortNumber() != registryContainerPort {
			continue
		}

		p.ExposedTo = &port

		return
	}
}

// overlayRegistryAuthEnv sets the registry:2 image's htpasswd basic-auth env
// vars - builtin/registry/deployment.yaml deliberately carries none of these
// (no descriptor file ever contains a credential).
func overlayRegistryAuthEnv(request *velez_api.CreateSmerd_Request) {
	if request.Env == nil {
		request.Env = make(map[string]string, 3) //nolint:mnd
	}

	request.Env[envRegistryAuth] = "htpasswd"
	request.Env[envRegistryAuthHtpasswdRealm] = registryAuthRealm
	request.Env[envRegistryAuthHtpasswdPath] = registryHtpasswdPath
}

// registerRegistryPluginJob mirrors enable_statefull.go's registerPluginJob:
// records the registry in velez.plugins so ListPlugins (and the frontend's
// plugin header) sees it. The velez.services/deployment_specifications/
// deployments rows are already handled by deployRegistryJob's
// CreateNewDeploy call, unlike enable_statefull.go's raw-Docker-created
// sidecar. No Rollback: this job only records state earlier jobs already
// established.
type registerRegistryPluginJob struct {
	storageContainer storage.Storage
}

func (j *registerRegistryPluginJob) Do(ctx context.Context) error {
	svc, err := j.storageContainer.Services().GetByName(ctx, RegistryServiceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting registry service")
	}

	pluginParams := plugins_queries.UpsertPluginParams{
		PluginType: velez_api.VervPluginType_registry.String(),
		ServiceID:  sql.NullInt64{Int64: svc.ID, Valid: true},
	}

	err = j.storageContainer.Plugins().UpsertPlugin(ctx, pluginParams)
	if err != nil {
		return rerrors.Wrap(err, "error upserting registry plugin")
	}

	return nil
}

// registerRegistryRowJob upserts the velez.registries row that makes the
// registry immediately usable for image pulls/pushes - secret is always the
// secret ref string (plugin/registry/password), never the password value,
// per docs/features/pgaas_and_registry_plugin.md section 1. RegistryServiceName
// is globally unique (single registry per node), and velez.registries.name is
// UNIQUE, so upserting by name is sufficient for idempotency across a
// crash-resumed task. The upsert itself goes through
// registries.BuiltinRegistryUpserter rather than storage.RegistriesStorage's
// public Create/Update - single-node mode's in-memory backend otherwise
// rejects every write (user_errors.ErrRequiresStatefullMode), but this row is
// the registry plugin's own system bookkeeping, not user-facing registry
// CRUD.
type registerRegistryRowJob struct {
	registries storage.RegistriesStorage

	ctx interface {
		registryUsernameAccessor
		registryPortAccessor
	}
}

func (j *registerRegistryRowJob) Do(ctx context.Context) error {
	upserter, ok := j.registries.(registries.BuiltinRegistryUpserter)
	if !ok {
		return rerrors.Wrap(user_errors.ErrRegistriesStorageMissingBuiltinUpsert)
	}

	req := domain.CreateRegistryReq{
		Name:     RegistryServiceName,
		Type:     domain.RegistryTypeGenericV2,
		Url:      registryUrl(j.ctx.GetExposedPort()),
		Username: j.ctx.GetUsername(),
		Secret:   registryPasswordSecretRef().String(),
	}

	_, err := upserter.UpsertBuiltinRegistry(ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error upserting registry row")
	}

	return nil
}

// registryUrl mirrors getRootDsnJob's applyBareBinaryHostPort split
// (enable_statefull.go): a Velez running inside a container reaches the
// registry over the Docker network by its container/service name on its own
// internal listening port; a Velez running as a bare binary can only reach
// it via the host-published port.
func registryUrl(exposedPort uint32) string {
	if env.IsInContainer() {
		hostPort := net.JoinHostPort(RegistryServiceName, strconv.Itoa(registryContainerPort))

		return "http://" + hostPort
	}

	hostPort := net.JoinHostPort("localhost", strconv.Itoa(int(exposedPort)))

	return "http://" + hostPort
}
