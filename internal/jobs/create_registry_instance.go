package jobs

import (
	"context"
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
	"go.vervstack.ru/Velez/internal/domain/labels"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/registries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	CreateRegistryInstanceAction = "create_registry_instance"

	// registryaasDefaultUsername is the fixed basic-auth username every
	// registry instance is provisioned with - CreateRegistryInstance.Request
	// carries no username field (unlike the retired EnableRegistry.Request),
	// so there is no per-request override path.
	registryaasDefaultUsername = "verv"

	// registryaasDescriptorName / registryaasUiDescriptorName select the
	// builtin descriptors - see
	// internal/service/service_manager/vervonomicon/builtin.
	registryaasDescriptorName   = "registry"
	registryaasUiDescriptorName = "registry_ui"

	// registryaasUiServiceSuffix names the UI sidecar's own service/container,
	// derived from the instance name - so multiple instances' sidecars never
	// collide in Docker's global container-name namespace.
	registryaasUiServiceSuffix = "-ui"

	// registryaasDataVolumeSuffix / registryaasAuthVolumeSuffix derive
	// per-instance Docker volume names from the instance name, mirroring
	// pgaas.pgVolumeName - multiple instances launched from the same builtin
	// descriptor (whose volume names are fixed to "registry-data"/
	// "registry-auth") would otherwise collide in Docker's global volume
	// namespace.
	registryaasDataVolumeSuffix = "-data"
	registryaasAuthVolumeSuffix = "-auth"

	registryaasHtpasswdPath = "/auth/htpasswd"

	// registryaasContainerPort / registryaasUiContainerPort are the two
	// images' own listening ports (builtin/registry and builtin/registry_ui
	// deployment.yaml's app.ports[0].port) - never the host-exposed ports,
	// which are resolved per instance by resolveRegistryPortsJob.
	registryaasContainerPort   = 5000
	registryaasUiContainerPort = 80

	// envRegistryAuth* overlay the registry:2 image's htpasswd basic-auth env
	// vars - builtin/registry/deployment.yaml deliberately carries no env
	// (per docs/features/vervonomicon.md, no descriptor file ever contains a
	// credential), so this job chain overlays them at deploy time instead.
	envRegistryAuth              = "REGISTRY_AUTH"
	envRegistryAuthHtpasswdRealm = "REGISTRY_AUTH_HTPASSWD_REALM"
	envRegistryAuthHtpasswdPath  = "REGISTRY_AUTH_HTPASSWD_PATH"
	registryaasAuthRealm         = "Registry Realm"

	// envNginxProxyPassUrl / envRegistryTitle overlay the docker-registry-ui
	// image's own env - see builtin/registry_ui/deployment.yaml's doc comment
	// on why these are instance-specific and resolved here rather than baked
	// into the descriptor.
	envNginxProxyPassUrl = "NGINX_PROXY_PASS_URL"
	envRegistryTitle     = "REGISTRY_TITLE"

	registryaasSecretScope = "registryaas"
	registryaasSecretKey   = "password"

	stepPutSecret                = "put_secret"
	stepResolvePorts             = "resolve_ports"
	stepWriteHtpasswd            = "write_htpasswd"
	stepDeployRegistry           = "deploy_registry"
	stepDeployRegistryUi         = "deploy_registry_ui"
	stepRegisterRegistryInstance = "register_registry_instance_row"
	stepRegisterRegistryRow      = "register_registry_row"
	stepBindOwnerResource        = "bind_owner_resource"

	// registryaasResourceType is the velez.service_resources.resource_type
	// stamped onto the (owner, instance) binding bindRegistryOwnerResourceJob
	// records when the request names an owner service - mirrors pgaas's
	// pgResourceType.
	registryaasResourceType = "container_registry"
)

// Accessor interfaces the create_registry_instance jobs need from their
// TaskContext. *velez_api.CreateRegistryInstanceTaskPayload satisfies all of
// them. containerIDAccessor is declared in create_smerd.go and reused here,
// exactly as copy_to_volume.go and the retired enable_registry.go did.

type createRegistryInstanceRequestAccessor interface {
	GetRequest() *velez_api.CreateRegistryInstance_Request
}

type registryUsernameAccessor interface {
	GetUsername() string
	SetUsername(username string)
}

type registryPasswordAccessor interface {
	GetPassword() string
	SetPassword(password string)
}

type registryExposedPortAccessor interface {
	GetExposedPort() uint32
	SetExposedPort(port uint32)
}

type registryUiExposedPortAccessor interface {
	GetUiExposedPort() uint32
	SetUiExposedPort(port uint32)
}

// registryPortManager is the narrow node_clients.PortManager slice
// resolveRegistryPortsJob needs - see that job's doc comment for why an
// auto-assigned port is put on hold. Declared narrow, following
// copy_to_volume.go's startAPI/copyAPI precedent, so this job stays
// hand-fakeable without depending on the full PortManager interface.
type registryPortManager interface {
	GetPortForEnvironment(environment string) (uint32, error)
	HoldPort(port uint32) bool
}

type createRegistryInstanceHandler struct {
	nodeClients node_clients.NodeClients
	runtimes    container_runtime.RuntimeResolver
	// storageContainer must be the same storage.Storage instance
	// VervServicesService.CreateNewDeploy reads/writes - see
	// enable_registry.go's now-retired doc comment on this same field for the
	// full rationale (unchanged: pre/post Postgres-cluster-join convergence).
	storageContainer storage.Storage
	secretsStore     secrets.Store
	vervServices     service.VervServicesService
}

func NewCreateRegistryInstanceHandler(
	nodeClients node_clients.NodeClients,
	runtimes container_runtime.RuntimeResolver,
	storageContainer storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
) TaskHandler {
	return &createRegistryInstanceHandler{
		nodeClients:      nodeClients,
		runtimes:         runtimes,
		storageContainer: storageContainer,
		secretsStore:     secretsStore,
		vervServices:     vervServices,
	}
}

func (h *createRegistryInstanceHandler) Action() string {
	return CreateRegistryInstanceAction
}

func (h *createRegistryInstanceHandler) NewContext() TaskContext {
	return &velez_api.CreateRegistryInstanceTaskPayload{}
}

// BuildJobs, per docs/features/pgaas_and_registry_plugin.md section 4 (as
// extended for Container-Registry-as-a-Service): generate_credentials/
// put_secret persist a username+password the same way enable_statefull.go's
// generateCredentialsJob does; resolve_ports resolves BOTH the registry's and
// the UI sidecar's host ports up front - Docker container labels are set at
// create time and can never be added later, and the single-node/dev
// local_storage backend recovers ui_port from a label on the registry
// container itself (internal/storage/local_storage/registry_instances.go),
// so that label must be known before deploy_registry runs, not resolved
// lazily inside it the way the retired enable_registry.go resolved its one
// port; the four create/start/write/drop steps write a bcrypt htpasswd entry
// into the instance's own registry-auth volume by reusing copy_to_volume.go's
// loader-container jobs verbatim; deploy_registry and deploy_registry_ui each
// resolve a builtin descriptor and hand off to the existing CreateNewDeploy
// path - the deploy watcher creates the actual containers, this handler never
// does; register_registry_instance_row/register_registry_row record
// velez.registry_instances and velez.registries (never velez.plugins - this
// is a first-class, multi-instance service, not a node-singleton plugin);
// bind_owner_resource only runs when the request named an owner service.
func (h *createRegistryInstanceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.CreateRegistryInstanceTaskPayload)
	if !ok {
		panic("create_registry_instance: BuildJobs called with mismatched TaskContext type")
	}

	instanceName := payload.GetRequest().GetName()

	folders := mountedFolders(registryaasAuthVolumeName(instanceName), []string{registryaasHtpasswdPath})

	namedJobs := []NamedJob{
		{
			Name: stepGenerateCredentials,
			Job: &generateRegistryCredentialsJob{
				ctx: payload,
			},
		},
		{
			Name: stepPutSecret,
			Job: &putRegistrySecretJob{
				secrets:      h.secretsStore,
				instanceName: instanceName,
				ctx:          payload,
			},
		},
		{
			Name: stepResolvePorts,
			Job: &resolveRegistryPortsJob{
				portManager: h.nodeClients.PortManager(),
				req:         payload,
				ctx:         payload,
			},
		},
		{
			Name: stepCreateLoaderContainer,
			Job: &createLoaderContainerJob{
				nodeClients: h.nodeClients,
				req:         registryInstanceVolumeRef{instanceName: instanceName},
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
				filePath: registryaasHtpasswdPath,
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
			Job: &deployRegistryInstanceJob{
				boxes:        h.storageContainer.ResourceBoxes(),
				vervServices: h.vervServices,
				req:          payload,
				ctx:          payload,
			},
		},
		{
			Name: stepDeployRegistryUi,
			Job: &deployRegistryUiJob{
				boxes:        h.storageContainer.ResourceBoxes(),
				vervServices: h.vervServices,
				req:          payload,
				ctx:          payload,
			},
		},
		{
			Name: stepRegisterRegistryInstance,
			Job: &registerRegistryInstanceRowJob{
				services:          h.storageContainer.Services(),
				registryInstances: h.storageContainer.RegistryInstances(),
				instanceName:      instanceName,
				ctx:               payload,
			},
		},
		{
			Name: stepRegisterRegistryRow,
			Job: &registerRegistryRowJob{
				registries:   h.storageContainer.Registries(),
				instanceName: instanceName,
				ctx:          payload,
			},
		},
	}

	if payload.GetRequest().GetOwnerService() != "" {
		bindJob := &bindRegistryOwnerResourceJob{
			serviceResources: h.storageContainer.ServiceResources(),
			ownerService:     payload.GetRequest().GetOwnerService(),
			instanceName:     instanceName,
		}

		namedJobs = append(namedJobs, NamedJob{Name: stepBindOwnerResource, Job: bindJob})
	}

	return namedJobs
}

// registryaasDataVolumeName / registryaasAuthVolumeName derive per-instance
// Docker volume names, mirroring pgaas.pgVolumeName.
func registryaasDataVolumeName(instanceName string) string {
	return instanceName + registryaasDataVolumeSuffix
}

func registryaasAuthVolumeName(instanceName string) string {
	return instanceName + registryaasAuthVolumeSuffix
}

// registryaasUiServiceName derives the UI sidecar's own service/container
// name from the instance name.
func registryaasUiServiceName(instanceName string) string {
	return instanceName + registryaasUiServiceSuffix
}

// registryInstanceVolumeRef adapts an instance's own auth volume name to
// copyToVolumeRequestAccessor (declared in copy_to_volume.go) so
// createLoaderContainerJob can be reused verbatim for the htpasswd loader
// container - GetPathToFiles is never called by createLoaderContainerJob.Do,
// only GetVolumeName is, so a nil map here is fine.
type registryInstanceVolumeRef struct {
	instanceName string
}

func (r registryInstanceVolumeRef) GetVolumeName() string {
	return registryaasAuthVolumeName(r.instanceName)
}

func (registryInstanceVolumeRef) GetPathToFiles() map[string][]byte {
	return nil
}

// generateRegistryCredentialsJob mirrors enable_statefull.go's
// generateCredentialsJob: derives the instance's basic-auth username/password
// once and persists them into the task's context, so a crash-resumed task
// reuses the same credentials instead of regenerating ones that would
// mismatch the htpasswd entry already written into the instance's auth
// volume. Username is always registryaasDefaultUsername - the wire contract
// carries no per-request override (see that constant's doc comment).
type generateRegistryCredentialsJob struct {
	ctx interface {
		registryUsernameAccessor
		registryPasswordAccessor
	}
}

func (j *generateRegistryCredentialsJob) Do(_ context.Context) error {
	if j.ctx.GetUsername() == "" {
		j.ctx.SetUsername(registryaasDefaultUsername)
	}

	if j.ctx.GetPassword() == "" {
		j.ctx.SetPassword(string(toolbox.RandomBase64(generatedPwdLength)))
	}

	return nil
}

// putRegistrySecretJob persists the generated password into the secret store
// under the registryaas/<instance>/password ref. No Rollback: the secret is
// keyed by a fixed ref, so a later Put (from a resumed task) simply
// overwrites it rather than needing to be undone.
type putRegistrySecretJob struct {
	secrets      secrets.Store
	instanceName string

	ctx registryPasswordAccessor
}

func (j *putRegistrySecretJob) Do(ctx context.Context) error {
	ref := registryInstanceSecretRef(j.instanceName)

	err := j.secrets.Put(ctx, ref, j.ctx.GetPassword())
	if err != nil {
		return rerrors.Wrap(err, "error putting registry instance password secret")
	}

	return nil
}

func registryInstanceSecretRef(instanceName string) domain.SecretRef {
	return domain.SecretRef{
		Scope: registryaasSecretScope,
		Owner: instanceName,
		Key:   registryaasSecretKey,
	}
}

// resolveRegistryPortsJob resolves both the registry's and the UI sidecar's
// host ports before either container is deployed - see BuildJobs's doc
// comment on why the ordering matters. The registry's port is the caller's
// requested expose_to_port verbatim when given (create_smerd's own
// lockPorts job locks it for real once the deploy watcher runs create_smerd),
// or auto-assigned from the node's shared port pool. The UI sidecar's port
// has no request-level override at all - it is always auto-assigned. Every
// auto-assigned port is put on hold (HoldPort) rather than left bare:
// create_smerd's lockPorts job runs against the same port later (it carries
// a non-nil ExposedTo via the descriptor-level ExposeTo overlay in
// buildRegistryDeployRequest/buildRegistryUiDeployRequest) and, seeing an
// already-locked port with no hold recorded, would call
// LockPortForEnvironment against a port it doesn't own yet and fail with
// ErrPortAlreadyLocked - LockPortForEnvironment has no "already locked by
// this same environment" no-op path, it always rejects a locked port. Holding
// it here makes lockPorts's UnHoldPort check consume the hold and skip
// re-locking instead. Mirrors the retired enable_registry.go's
// deployRegistryJob.resolvePort, extended to two ports.
type resolveRegistryPortsJob struct {
	portManager registryPortManager

	req createRegistryInstanceRequestAccessor
	ctx interface {
		registryExposedPortAccessor
		registryUiExposedPortAccessor
	}
}

func (j *resolveRegistryPortsJob) Do(_ context.Context) error {
	if j.ctx.GetExposedPort() == 0 {
		port, err := j.resolveRegistryPort()
		if err != nil {
			return err
		}

		j.ctx.SetExposedPort(port)
	}

	if j.ctx.GetUiExposedPort() == 0 {
		environment := j.req.GetRequest().GetEnvironment()

		port, err := j.portManager.GetPortForEnvironment(environment)
		if err != nil {
			return rerrors.Wrap(err, "error resolving free port for registry ui sidecar")
		}

		j.portManager.HoldPort(port)
		j.ctx.SetUiExposedPort(port)
	}

	return nil
}

func (j *resolveRegistryPortsJob) resolveRegistryPort() (uint32, error) {
	requested := j.req.GetRequest().GetExposeToPort()
	if requested != 0 {
		return requested, nil
	}

	environment := j.req.GetRequest().GetEnvironment()

	port, err := j.portManager.GetPortForEnvironment(environment)
	if err != nil {
		return 0, rerrors.Wrap(err, "error resolving free port for registry")
	}

	j.portManager.HoldPort(port)

	return port, nil
}

// writeHtpasswdJob writes the bcrypt htpasswd entry into the loader
// container created by createLoaderContainerJob, reusing
// copy_to_volume.go's writeFileToContainer tar-write helper verbatim (same
// mkdir-then-write shape as copyFileJob.Do). It can't reuse copyFileJob
// itself: copyFileJob.content is a plain []byte fixed at BuildJobs time,
// before generate_credentials has run, so it can never carry a
// generated-this-run password.
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

// deployRegistryInstanceJob resolves the builtin registry descriptor,
// overlays the instance's per-instance volume names, resolved port, auth env
// vars and discovery labels, and hands off to the same
// VervServicesService.CreateNewDeploy path any other vervonomicon service
// deploys through - the deploy watcher creates and starts the actual
// container from there. This job never creates a container itself.
type deployRegistryInstanceJob struct {
	boxes        vervonomicon.BoxLookup
	vervServices service.VervServicesService

	req createRegistryInstanceRequestAccessor
	ctx interface {
		registryUsernameAccessor
		registryExposedPortAccessor
		registryUiExposedPortAccessor
	}
}

func (j *deployRegistryInstanceJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()
	instanceName := request.GetName()

	descriptor, smerdRequest, err := buildRegistryDeployRequest(ctx, j.boxes, request, j.ctx.GetExposedPort())
	if err != nil {
		return rerrors.Wrap(err, "error building registry instance deploy request")
	}

	overlayRegistryAuthEnv(smerdRequest)

	if smerdRequest.Labels == nil {
		smerdRequest.Labels = make(map[string]string)
	}

	smerdRequest.Labels[labels.VervServiceLabel] = instanceName
	smerdRequest.Labels[labels.RegistryaasInstanceLabel] = "true"
	smerdRequest.Labels[labels.RegistryaasUsernameLabel] = j.ctx.GetUsername()
	smerdRequest.Labels[labels.RegistryaasUiPortLabel] = strconv.Itoa(int(j.ctx.GetUiExposedPort()))

	deployReq := domain.CreateDeployReq{
		ServiceName:    instanceName,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating registry instance deploy")
	}

	return nil
}

// buildRegistryDeployRequest reads the builtin registry descriptor, overlays
// the instance's shape (box, unique volume names, exposed port), and
// resolves it into a CreateSmerd.Request via the box resolver - mirrors
// pgaas.buildDeployRequest.
func buildRegistryDeployRequest(
	ctx context.Context,
	boxes vervonomicon.BoxLookup,
	request *velez_api.CreateRegistryInstance_Request,
	exposedPort uint32,
) (verv.Descriptor, *velez_api.CreateSmerd_Request, error) {
	files, err := builtin.Read(registryaasDescriptorName)
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error reading builtin registry descriptor")
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, request.GetEnvironment())
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error merging registry vervonomicon environment overlay")
	}

	descriptor.Source = verv.SourceKindBuiltin

	if request.GetBox() != "" {
		descriptor.Deployment.App.Box = request.GetBox()
	}

	instanceName := request.GetName()

	for i := range descriptor.Deployment.App.Volumes {
		switch descriptor.Deployment.App.Volumes[i].Name {
		case "registry-data":
			descriptor.Deployment.App.Volumes[i].Name = registryaasDataVolumeName(instanceName)
		case "registry-auth":
			descriptor.Deployment.App.Volumes[i].Name = registryaasAuthVolumeName(instanceName)
		}
	}

	if len(descriptor.Deployment.App.Ports) > 0 {
		descriptor.Deployment.App.Ports[0].ExposeTo = int(exposedPort)
	}

	resolver := vervonomicon.NewBoxResolver(boxes)

	smerdRequest, err := resolver.ResolveRequest(ctx, descriptor, request.GetEnvironment(), "")
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error resolving registry instance deploy request")
	}

	smerdRequest.Name = instanceName

	return descriptor, smerdRequest, nil
}

// overlayRegistryAuthEnv sets the registry:2 image's htpasswd basic-auth env
// vars - builtin/registry/deployment.yaml deliberately carries none of these
// (no descriptor file ever contains a credential).
func overlayRegistryAuthEnv(request *velez_api.CreateSmerd_Request) {
	if request.Env == nil {
		request.Env = make(map[string]string, 3) //nolint:mnd
	}

	request.Env[envRegistryAuth] = "htpasswd"
	request.Env[envRegistryAuthHtpasswdRealm] = registryaasAuthRealm
	request.Env[envRegistryAuthHtpasswdPath] = registryaasHtpasswdPath
}

// deployRegistryUiJob resolves the builtin registry_ui descriptor, overlays
// the sidecar's proxy target/title and resolved port, and hands off to the
// same CreateNewDeploy path deployRegistryInstanceJob uses - a distinct
// service/container from the registry instance itself, named
// registryaasUiServiceName(instanceName).
type deployRegistryUiJob struct {
	boxes        vervonomicon.BoxLookup
	vervServices service.VervServicesService

	req createRegistryInstanceRequestAccessor
	ctx registryUiExposedPortAccessor
}

func (j *deployRegistryUiJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()
	instanceName := request.GetName()
	uiServiceName := registryaasUiServiceName(instanceName)

	files, err := builtin.Read(registryaasUiDescriptorName)
	if err != nil {
		return rerrors.Wrap(err, "error reading builtin registry ui descriptor")
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, request.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error merging registry ui vervonomicon environment overlay")
	}

	descriptor.Source = verv.SourceKindBuiltin

	if len(descriptor.Deployment.App.Ports) > 0 {
		descriptor.Deployment.App.Ports[0].ExposeTo = int(j.ctx.GetUiExposedPort())
	}

	resolver := vervonomicon.NewBoxResolver(j.boxes)

	smerdRequest, err := resolver.ResolveRequest(ctx, descriptor, request.GetEnvironment(), "")
	if err != nil {
		return rerrors.Wrap(err, "error resolving registry ui deploy request")
	}

	smerdRequest.Name = uiServiceName

	if smerdRequest.Env == nil {
		smerdRequest.Env = make(map[string]string, 2) //nolint:mnd
	}

	smerdRequest.Env[envNginxProxyPassUrl] = registryInternalUrl(instanceName)
	smerdRequest.Env[envRegistryTitle] = instanceName

	if smerdRequest.Labels == nil {
		smerdRequest.Labels = make(map[string]string)
	}

	smerdRequest.Labels[labels.VervServiceLabel] = uiServiceName

	deployReq := domain.CreateDeployReq{
		ServiceName:    uiServiceName,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating registry ui deploy")
	}

	return nil
}

// registryInternalUrl is the registry instance's own address on the Docker
// network - always reachable by its container/service name, regardless of
// whether Velez itself runs inside a container (unlike registryInstanceUrl,
// which is Velez's own view from outside).
func registryInternalUrl(instanceName string) string {
	hostPort := net.JoinHostPort(instanceName, strconv.Itoa(registryaasContainerPort))

	return "http://" + hostPort
}

// registerRegistryInstanceRowJob upserts the velez.registry_instances row -
// the registry-specific facts (port, ui_port, username, secret_ref) linked to
// the service CreateNewDeploy already created. Status/environment/image are
// never duplicated here - they're read back through VervServicesService.
type registerRegistryInstanceRowJob struct {
	services          storage.ServicesStorage
	registryInstances storage.RegistryInstancesStorage
	instanceName      string

	ctx interface {
		registryUsernameAccessor
		registryUiExposedPortAccessor
	}
}

func (j *registerRegistryInstanceRowJob) Do(ctx context.Context) error {
	svc, err := j.services.GetByName(ctx, j.instanceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting registry instance service")
	}

	secretRef := registryInstanceSecretRef(j.instanceName)

	upsertReq := domain.UpsertRegistryInstanceReq{
		ServiceId: svc.ID,
		Port:      registryaasContainerPort,
		UiPort:    int32(j.ctx.GetUiExposedPort()), //nolint:gosec
		Username:  j.ctx.GetUsername(),
		SecretRef: secretRef.String(),
	}

	_, err = j.registryInstances.UpsertRegistryInstance(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting registry instance row")
	}

	return nil
}

// registerRegistryRowJob upserts the velez.registries row that makes the
// instance immediately usable for image pulls/pushes - secret is always the
// secret ref string (registryaas/<instance>/password), never the password
// value. The upsert itself goes through registries.BuiltinRegistryUpserter
// rather than storage.RegistriesStorage's public Create/Update - single-node
// mode's in-memory backend otherwise rejects every write
// (user_errors.ErrRequiresStatefullMode), but this row is the instance's own
// system bookkeeping, not user-facing registry CRUD. Mirrors the retired
// enable_registry.go's registerRegistryRowJob.
type registerRegistryRowJob struct {
	registries   storage.RegistriesStorage
	instanceName string

	ctx interface {
		registryUsernameAccessor
		registryExposedPortAccessor
	}
}

func (j *registerRegistryRowJob) Do(ctx context.Context) error {
	upserter, ok := j.registries.(registries.BuiltinRegistryUpserter)
	if !ok {
		return rerrors.Wrap(user_errors.ErrRegistriesStorageMissingBuiltinUpsert)
	}

	secretRef := registryInstanceSecretRef(j.instanceName)

	req := domain.CreateRegistryReq{
		Name:     j.instanceName,
		Type:     domain.RegistryTypeGenericV2,
		Url:      registryInstanceUrl(j.instanceName, j.ctx.GetExposedPort()),
		Username: j.ctx.GetUsername(),
		Secret:   secretRef.String(),
	}

	_, err := upserter.UpsertBuiltinRegistry(ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error upserting registry row")
	}

	return nil
}

// registryInstanceUrl mirrors the retired enable_registry.go's registryUrl,
// parameterized on the instance name: a Velez running inside a container
// reaches the instance over the Docker network by its container/service name
// on its own internal listening port; a Velez running as a bare binary can
// only reach it via the host-published port.
func registryInstanceUrl(instanceName string, exposedPort uint32) string {
	if env.IsInContainer() {
		return registryInternalUrl(instanceName)
	}

	hostPort := net.JoinHostPort("localhost", strconv.Itoa(int(exposedPort)))

	return "http://" + hostPort
}

// bindRegistryOwnerResourceJob records the velez.service_resources binding
// making the instance a bound resource of its owner service, once ownerService
// is non-empty - mirrors pgaas's CreatePgInstance owner-binding step. No
// Rollback: this job only records state earlier jobs already established.
type bindRegistryOwnerResourceJob struct {
	serviceResources storage.ServiceResourcesStorage
	ownerService     string
	instanceName     string
}

func (j *bindRegistryOwnerResourceJob) Do(ctx context.Context) error {
	err := j.serviceResources.UpsertResource(ctx, j.ownerService, j.instanceName, registryaasResourceType)
	if err != nil {
		return rerrors.Wrap(err, "error binding registry instance to owner service")
	}

	return nil
}
