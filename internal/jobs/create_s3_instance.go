package jobs

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"maps"
	"net"
	"strconv"
	"strings"
	"time"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/garage"
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
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	CreateS3InstanceAction = "create_s3_instance"

	stepGenerateS3Secrets = "generate_secrets"
	stepDeployGarage      = "deploy_garage"
	stepWaitGarage        = "wait_garage"
	stepApplyGarageLayout = "apply_layout"
	stepDeployGarageWebUi = "deploy_web_ui"
	stepWaitGarageWebUi   = "wait_web_ui"

	s3GarageDescriptorName = "garage"
	s3WebUiDescriptorName  = "garage-webui"

	s3MetaDescriptorVolume = "garage-meta"
	s3DataDescriptorVolume = "garage-data"

	envGarageRpcSecret  = "GARAGE_RPC_SECRET"
	envGarageAdminToken = "GARAGE_ADMIN_TOKEN"

	envWebUiApiBaseUrl    = "API_BASE_URL"
	envWebUiS3EndpointUrl = "S3_ENDPOINT_URL"
	envWebUiS3Region      = "S3_REGION"
	envWebUiApiAdminKey   = "API_ADMIN_KEY"
	envWebUiAuthUserPass  = "AUTH_USER_PASS"

	s3RpcSecretByteCount     = 32
	s3AdminTokenByteCount    = 32
	s3WebUiPasswordByteCount = 12

	s3DeployWaitTimeout  = 180 * time.Second
	s3HealthWaitTimeout  = 60 * time.Second
	s3HealthPollInterval = time.Second

	s3LayoutZone                  = "dc1"
	s3LayoutNominalCapacity int64 = 1_000_000_000

	garageConfigTemplate = `metadata_dir = "%s"
data_dir = "%s"
db_engine = "sqlite"
replication_factor = 1
rpc_bind_addr = "[::]:%d"
rpc_public_addr = "%s:%d"

[s3_api]
s3_region = "%s"
api_bind_addr = "[::]:%d"

[admin]
api_bind_addr = "[::]:%d"
`
)

var errGarageNoNodes = rerrors.New("garage cluster status lists no nodes")

type createS3InstanceHandler struct {
	nodeClients    node_clients.NodeClients
	dataStorage    storage.Storage
	secretsStore   secrets.Store
	vervServices   service.VervServicesService
	configResolver service.ServiceConfigResolver
	jobsEngine     Engine
	runtimes       container_runtime.RuntimeResolver
}

func NewCreateS3InstanceHandler(
	nodeClients node_clients.NodeClients,
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
	configResolver service.ServiceConfigResolver,
	jobsEngine Engine,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &createS3InstanceHandler{
		nodeClients:    nodeClients,
		dataStorage:    dataStorage,
		secretsStore:   secretsStore,
		vervServices:   vervServices,
		configResolver: configResolver,
		jobsEngine:     jobsEngine,
		runtimes:       runtimes,
	}
}

func (h *createS3InstanceHandler) Action() string {
	return CreateS3InstanceAction
}

func (h *createS3InstanceHandler) NewContext() TaskContext {
	return &velez_api.CreateS3InstanceTaskPayload{}
}

func (h *createS3InstanceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.CreateS3InstanceTaskPayload)
	if !ok {
		panic("create_s3_instance: BuildJobs called with mismatched TaskContext type")
	}

	namedJobs := []NamedJob{
		{
			Name: stepGenerateS3Secrets,
			Job:  &generateS3SecretsJob{secrets: h.secretsStore, payload: payload},
		},
		{
			Name: stepResolvePorts,
			Job:  &resolveS3PortsJob{portManager: h.nodeClients.PortManager(), payload: payload},
		},
		{
			Name: stepDeployGarage,
			Job: &deployGarageJob{
				boxes:          h.dataStorage.ResourceBoxes(),
				vervServices:   h.vervServices,
				configResolver: h.configResolver,
				secrets:        h.secretsStore,
				payload:        payload,
			},
		},
		{
			Name: stepWaitGarage,
			Job: &waitGarageJob{
				jobsEngine: h.jobsEngine,
				payload:    payload,
			},
		},
		{
			Name: stepApplyGarageLayout,
			Job: &applyGarageLayoutJob{
				runtimes: h.runtimes,
				secrets:  h.secretsStore,
				payload:  payload,
			},
		},
	}

	if payload.GetRequest().GetEnableWebUi() {
		deployWebUi := &deployGarageWebUiJob{
			boxes:          h.dataStorage.ResourceBoxes(),
			vervServices:   h.vervServices,
			configResolver: h.configResolver,
			secrets:        h.secretsStore,
			payload:        payload,
		}

		waitWebUi := &waitGarageWebUiJob{jobsEngine: h.jobsEngine, payload: payload}

		namedJobs = append(namedJobs,
			NamedJob{Name: stepDeployGarageWebUi, Job: deployWebUi},
			NamedJob{Name: stepWaitGarageWebUi, Job: waitWebUi},
		)
	}

	return namedJobs
}

func s3Region(request *velez_api.CreateS3Instance_Request) string {
	return toolbox.Coalesce(request.GetRegion(), domain.S3DefaultRegion)
}

func s3RandomHex(byteCount int) (string, error) {
	buf := make([]byte, byteCount)

	_, err := rand.Read(buf)
	if err != nil {
		return "", rerrors.Wrap(err, "error reading random bytes")
	}

	return hex.EncodeToString(buf), nil
}

func s3RandomBase64(byteCount int) (string, error) {
	buf := make([]byte, byteCount)

	_, err := rand.Read(buf)
	if err != nil {
		return "", rerrors.Wrap(err, "error reading random bytes")
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

// ensureSecret keeps a resumed task on the secret an earlier run already
// stored: a regenerated rpc secret or admin token would not match the
// container that was already deployed with the old one.
func ensureSecret(
	ctx context.Context,
	store secrets.Store,
	ref domain.SecretRef,
	generate func() (string, error),
) error {
	_, err := store.Get(ctx, ref)
	if err == nil {
		return nil
	}

	if !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error reading secret")
	}

	value, err := generate()
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = store.Put(ctx, ref, value)
	if err != nil {
		return rerrors.Wrap(err, "error putting secret")
	}

	return nil
}

type generateS3SecretsJob struct {
	secrets secrets.Store
	payload *velez_api.CreateS3InstanceTaskPayload
}

func (j *generateS3SecretsJob) Do(ctx context.Context) error {
	name := j.payload.GetRequest().GetName()

	err := ensureSecret(ctx, j.secrets, domain.S3RpcSecretRef(name), generateS3RpcSecret)
	if err != nil {
		return rerrors.Wrap(err, "error ensuring rpc secret")
	}

	err = ensureSecret(ctx, j.secrets, domain.S3AdminTokenSecretRef(name), generateS3AdminToken)
	if err != nil {
		return rerrors.Wrap(err, "error ensuring admin token")
	}

	if !j.payload.GetRequest().GetEnableWebUi() {
		return nil
	}

	err = ensureSecret(ctx, j.secrets, domain.S3WebUiPasswordSecretRef(name), generateS3WebUiPassword)
	if err != nil {
		return rerrors.Wrap(err, "error ensuring web ui password")
	}

	return nil
}

func generateS3RpcSecret() (string, error) {
	return s3RandomHex(s3RpcSecretByteCount)
}

func generateS3AdminToken() (string, error) {
	return s3RandomBase64(s3AdminTokenByteCount)
}

func generateS3WebUiPassword() (string, error) {
	return s3RandomHex(s3WebUiPasswordByteCount)
}

// resolveS3PortsJob holds every auto-assigned port for the same reason
// resolveRegistryPortsJob does - see its doc comment.
type resolveS3PortsJob struct {
	portManager registryPortManager
	payload     *velez_api.CreateS3InstanceTaskPayload
}

func (j *resolveS3PortsJob) Do(_ context.Context) error {
	request := j.payload.GetRequest()
	environment := request.GetEnvironment()

	if j.payload.ExposedPort == nil {
		port, err := j.s3Port(environment, request.GetExposeToPort())
		if err != nil {
			return err
		}

		j.payload.ExposedPort = &port
	}

	if request.GetEnableWebUi() && j.payload.WebUiExposedPort == nil {
		port, err := j.heldPort(environment)
		if err != nil {
			return err
		}

		j.payload.WebUiExposedPort = &port
	}

	if !env.IsInContainer() && j.payload.AdminExposedPort == nil {
		port, err := j.heldPort(environment)
		if err != nil {
			return err
		}

		j.payload.AdminExposedPort = &port
	}

	return nil
}

func (j *resolveS3PortsJob) s3Port(environment string, requested uint32) (uint32, error) {
	if requested != 0 {
		return requested, nil
	}

	return j.heldPort(environment)
}

func (j *resolveS3PortsJob) heldPort(environment string) (uint32, error) {
	port, err := j.portManager.GetPortForEnvironment(environment)
	if err != nil {
		return 0, rerrors.Wrap(err, "error resolving free port for s3 instance")
	}

	j.portManager.HoldPort(port)

	return port, nil
}

func exposeDescriptorPorts(descriptor *verv.Descriptor, hostPorts map[int]uint32) {
	ports := descriptor.Deployment.App.Ports

	for i := range ports {
		hostPort := hostPorts[ports[i].Port]
		if hostPort != 0 {
			ports[i].ExposeTo = int(hostPort)
		}
	}
}

func renameDescriptorVolumes(descriptor *verv.Descriptor, names map[string]string) {
	volumes := descriptor.Deployment.App.Volumes

	for i := range volumes {
		newName, ok := names[volumes[i].Name]
		if ok {
			volumes[i].Name = newName
		}
	}
}

func attachS3Network(request *velez_api.CreateSmerd_Request, instanceName string) {
	if request.GetSettings() == nil {
		request.Settings = &velez_api.Container_Settings{}
	}

	networkBind := &velez_api.NetworkBind{NetworkName: domain.S3NetworkName(instanceName)}

	request.Settings.Network = append(request.Settings.Network, networkBind)
}

func setRequestEnv(request *velez_api.CreateSmerd_Request, values map[string]string) {
	if request.Env == nil {
		request.Env = make(map[string]string, len(values))
	}

	maps.Copy(request.GetEnv(), values)
}

func resolveS3Descriptor(
	ctx context.Context,
	boxes vervonomicon.BoxLookup,
	descriptorName, environment, box string,
	hostPorts map[int]uint32,
	volumeNames map[string]string,
) (verv.Descriptor, *velez_api.CreateSmerd_Request, error) {
	files, err := builtin.Read(descriptorName)
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrapf(err, "error reading builtin %s descriptor", descriptorName)
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, environment)
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrapf(err, "error merging %s environment overlay", descriptorName)
	}

	descriptor.Source = verv.SourceKindBuiltin

	if box != "" {
		descriptor.Deployment.App.Box = box
	}

	exposeDescriptorPorts(&descriptor, hostPorts)
	renameDescriptorVolumes(&descriptor, volumeNames)

	resolver := vervonomicon.NewBoxResolver(boxes)

	smerdRequest, err := resolver.ResolveRequest(ctx, descriptor, environment, "")
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrapf(err, "error resolving %s deploy request", descriptorName)
	}

	return descriptor, smerdRequest, nil
}

func renderGarageConfig(instanceName, region string) string {
	return fmt.Sprintf(
		garageConfigTemplate,
		domain.S3MetaPath,
		domain.S3DataPath,
		domain.S3RpcContainerPort,
		instanceName,
		domain.S3RpcContainerPort,
		region,
		domain.S3ApiContainerPort,
		domain.S3AdminContainerPort,
	)
}

type deployGarageJob struct {
	boxes          vervonomicon.BoxLookup
	vervServices   service.VervServicesService
	configResolver service.ServiceConfigResolver
	secrets        secrets.Store
	payload        *velez_api.CreateS3InstanceTaskPayload
}

func (j *deployGarageJob) Do(ctx context.Context) error {
	request := j.payload.GetRequest()
	name := request.GetName()

	hostPorts := map[int]uint32{
		domain.S3ApiContainerPort:   j.payload.GetExposedPort(),
		domain.S3AdminContainerPort: j.payload.GetAdminExposedPort(),
	}

	volumeNames := map[string]string{
		s3MetaDescriptorVolume: domain.S3MetaVolumeName(name),
		s3DataDescriptorVolume: domain.S3DataVolumeName(name),
	}

	descriptor, smerdRequest, err := resolveS3Descriptor(
		ctx, j.boxes, s3GarageDescriptorName, request.GetEnvironment(), request.GetBox(), hostPorts, volumeNames,
	)
	if err != nil {
		return rerrors.Wrap(err, "error building garage deploy request")
	}

	serviceName := domain.S3ServiceName(name)

	smerdRequest.Name = serviceName

	if smerdRequest.Labels == nil {
		smerdRequest.Labels = make(map[string]string)
	}

	smerdRequest.Labels[labels.VervServiceLabel] = serviceName
	smerdRequest.Labels[labels.DisplayNameLabel] = name
	smerdRequest.Labels[labels.S3InstanceLabel] = name

	attachS3Network(smerdRequest, name)

	config := renderGarageConfig(serviceName, s3Region(request))

	err = j.configResolver.WriteFile(ctx, smerdRequest, domain.S3ConfigPath, []byte(config))
	if err != nil {
		return rerrors.Wrap(err, "error writing garage config")
	}

	rpcSecret, err := j.secrets.Get(ctx, domain.S3RpcSecretRef(name))
	if err != nil {
		return rerrors.Wrap(err, "error reading rpc secret")
	}

	adminToken, err := j.secrets.Get(ctx, domain.S3AdminTokenSecretRef(name))
	if err != nil {
		return rerrors.Wrap(err, "error reading admin token")
	}

	secretEnv := map[string]string{
		envGarageRpcSecret:  rpcSecret,
		envGarageAdminToken: adminToken,
	}

	setRequestEnv(smerdRequest, secretEnv)

	deployReq := domain.CreateDeployReq{
		ServiceName:    serviceName,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating garage deploy")
	}

	return nil
}

func waitForSmerdDeploy(
	ctx context.Context,
	jobsEngine taskWatcher,
	environment, serviceName string,
) error {
	entityID := SmerdEntityID(environment, serviceName)

	watchCtx, cancel := context.WithTimeout(ctx, s3DeployWaitTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range jobsEngine.Watch(watchCtx, entityID, CreateSmerdAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for %s to deploy, last status: %q",
			serviceName,
			finalTask.Status,
		)
	}

	if isFailed {
		return rerrors.Wrap(user_errors.ErrTaskFailed, finalTask.Error.String)
	}

	return nil
}

func newGarageClient(
	ctx context.Context,
	runtimes container_runtime.RuntimeResolver,
	store secrets.Store,
	environment, instanceName string,
) (*garage.Client, error) {
	adminToken, err := store.Get(ctx, domain.S3AdminTokenSecretRef(instanceName))
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading admin token")
	}

	containerRuntime, err := runtimes.Runtime(ctx, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	adminUrl, err := garage.AdminUrl(ctx, containerRuntime, domain.S3ServiceName(instanceName))
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving garage admin url")
	}

	return garage.New(adminUrl, adminToken), nil
}

// Garage answers /health with 503 until a layout assigns the node a role, so
// /health is only waited on after the layout is applied.
func waitGarageHealthy(ctx context.Context, client *garage.Client) error {
	return pollGarage(ctx, client.Health, "timed out waiting for garage to become healthy")
}

func pollGarage(ctx context.Context, probe func(context.Context) error, timeoutMessage string) error {
	pollCtx, cancel := context.WithTimeout(ctx, s3HealthWaitTimeout)
	defer cancel()

	ticker := time.NewTicker(s3HealthPollInterval)
	defer ticker.Stop()

	for {
		err := probe(pollCtx)
		if err == nil {
			return nil
		}

		select {
		case <-pollCtx.Done():
			return rerrors.Wrap(err, timeoutMessage)
		case <-ticker.C:
		}
	}
}

type waitGarageJob struct {
	jobsEngine taskWatcher
	payload    *velez_api.CreateS3InstanceTaskPayload
}

func (j *waitGarageJob) Do(ctx context.Context) error {
	request := j.payload.GetRequest()

	err := waitForSmerdDeploy(
		ctx, j.jobsEngine, request.GetEnvironment(), domain.S3ServiceName(request.GetName()),
	)
	if err != nil {
		return rerrors.Wrap(err, "error waiting for garage deploy")
	}

	return nil
}

type applyGarageLayoutJob struct {
	runtimes container_runtime.RuntimeResolver
	secrets  secrets.Store
	payload  *velez_api.CreateS3InstanceTaskPayload
}

func (j *applyGarageLayoutJob) Do(ctx context.Context) error {
	request := j.payload.GetRequest()

	client, err := newGarageClient(ctx, j.runtimes, j.secrets, request.GetEnvironment(), request.GetName())
	if err != nil {
		return err
	}

	status, err := client.GetClusterStatus(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error getting garage cluster status")
	}

	if len(status.Nodes) == 0 {
		return rerrors.Wrap(errGarageNoNodes)
	}

	node := status.Nodes[0]

	j.payload.NodeId = &node.Id

	if node.Role != nil {
		return waitGarageHealthy(ctx, client)
	}

	layout, err := client.GetClusterLayout(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error getting garage cluster layout")
	}

	// Capacity is only a relative weight between nodes; a single node ignores it.
	capacity := s3LayoutNominalCapacity

	role := garage.NodeRoleAssign{
		Id:       node.Id,
		Zone:     s3LayoutZone,
		Capacity: &capacity,
		Tags:     []string{},
	}

	roles := []garage.NodeRoleAssign{role}

	err = client.UpdateClusterLayout(ctx, roles)
	if err != nil {
		return rerrors.Wrap(err, "error staging garage layout")
	}

	err = client.ApplyClusterLayout(ctx, layout.Version+1)
	if err != nil {
		return rerrors.Wrap(err, "error applying garage layout")
	}

	return waitGarageHealthy(ctx, client)
}

type deployGarageWebUiJob struct {
	boxes          vervonomicon.BoxLookup
	vervServices   service.VervServicesService
	configResolver service.ServiceConfigResolver
	secrets        secrets.Store
	payload        *velez_api.CreateS3InstanceTaskPayload
}

func (j *deployGarageWebUiJob) Do(ctx context.Context) error {
	request := j.payload.GetRequest()
	name := request.GetName()
	serviceName := domain.S3ServiceName(name)
	webUiName := domain.S3WebUiServiceName(name)

	hostPorts := map[int]uint32{domain.S3WebUiContainerPort: j.payload.GetWebUiExposedPort()}

	descriptor, smerdRequest, err := resolveS3Descriptor(
		ctx, j.boxes, s3WebUiDescriptorName, request.GetEnvironment(), "", hostPorts, nil,
	)
	if err != nil {
		return rerrors.Wrap(err, "error building garage web ui deploy request")
	}

	smerdRequest.Name = webUiName

	if smerdRequest.Labels == nil {
		smerdRequest.Labels = make(map[string]string)
	}

	smerdRequest.Labels[labels.VervServiceLabel] = webUiName
	smerdRequest.Labels[labels.DisplayNameLabel] = name
	smerdRequest.Labels[labels.S3WebUiLabel] = name
	smerdRequest.Labels[labels.ComposeGroupLabel] = serviceName

	attachS3Network(smerdRequest, name)

	plainEnv := map[string]string{
		envWebUiApiBaseUrl:    s3InternalUrl(serviceName, domain.S3AdminContainerPort),
		envWebUiS3EndpointUrl: s3InternalUrl(serviceName, domain.S3ApiContainerPort),
		envWebUiS3Region:      s3Region(request),
	}

	err = j.configResolver.WriteEnv(ctx, smerdRequest, plainEnv)
	if err != nil {
		return rerrors.Wrap(err, "error writing garage web ui env")
	}

	secretEnv, err := j.secretEnv(ctx, name)
	if err != nil {
		return err
	}

	setRequestEnv(smerdRequest, secretEnv)

	deployReq := domain.CreateDeployReq{
		ServiceName:    webUiName,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating garage web ui deploy")
	}

	return nil
}

func (j *deployGarageWebUiJob) secretEnv(ctx context.Context, name string) (map[string]string, error) {
	adminToken, err := j.secrets.Get(ctx, domain.S3AdminTokenSecretRef(name))
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading admin token")
	}

	password, err := j.secrets.Get(ctx, domain.S3WebUiPasswordSecretRef(name))
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading web ui password")
	}

	line, err := htpasswdLine(domain.S3WebUiUsername, password)
	if err != nil {
		return nil, rerrors.Wrap(err, "error hashing web ui password")
	}

	secretEnv := map[string]string{
		envWebUiApiAdminKey:  adminToken,
		envWebUiAuthUserPass: strings.TrimSuffix(line, "\n"),
	}

	return secretEnv, nil
}

func s3InternalUrl(instanceName string, port int) string {
	hostPort := net.JoinHostPort(instanceName, strconv.Itoa(port))

	return "http://" + hostPort
}

type waitGarageWebUiJob struct {
	jobsEngine taskWatcher
	payload    *velez_api.CreateS3InstanceTaskPayload
}

func (j *waitGarageWebUiJob) Do(ctx context.Context) error {
	request := j.payload.GetRequest()
	webUiName := domain.S3WebUiServiceName(request.GetName())

	return waitForSmerdDeploy(ctx, j.jobsEngine, request.GetEnvironment(), webUiName)
}
