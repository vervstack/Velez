package jobs

import (
	"bytes"
	"context"
	"net"
	"regexp"
	"text/template"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	CreateHaproxyInstanceAction     = "create_haproxy_instance"
	stepValidateHaproxyMaster       = "validate_master"
	stepDeployHaproxy               = "deploy_haproxy"
	stepRecordHaproxyDeployBaseline = "record_deploy_baseline"
	stepWaitHaproxy                 = "wait_haproxy"
	stepAwaitHaproxyStable          = "await_stable"

	haproxyDescriptorName   = "haproxy"
	haproxyDescriptorVolume = "haproxy-config"
	haproxyConfigTemplate   = `defaults
    mode tcp
    timeout connect 5s
    timeout client 30s
    timeout server 30s

listen pg
    bind *:{{.Port}}
    option pgsql-check user {{.User}}
    server {{.Master}} {{.Address}} check
`
)

var (
	haproxyTemplate          = template.Must(template.New("haproxy.cfg").Parse(haproxyConfigTemplate))
	haproxyMasterNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)
)

type createHaproxyInstanceHandler struct {
	dataStorage    storage.Storage
	vervServices   service.VervServicesService
	configResolver service.ServiceConfigResolver
	jobsEngine     Engine
	runtimes       container_runtime.RuntimeResolver
	addressBook    service.AddressBook
}

func NewCreateHaproxyInstanceHandler(
	dataStorage storage.Storage,
	vervServices service.VervServicesService,
	configResolver service.ServiceConfigResolver,
	jobsEngine Engine,
	runtimes container_runtime.RuntimeResolver,
	addressBook service.AddressBook,
) TaskHandler {
	return &createHaproxyInstanceHandler{
		dataStorage:    dataStorage,
		vervServices:   vervServices,
		configResolver: configResolver,
		jobsEngine:     jobsEngine,
		runtimes:       runtimes,
		addressBook:    addressBook,
	}
}

func (h *createHaproxyInstanceHandler) Action() string {
	return CreateHaproxyInstanceAction
}

func (h *createHaproxyInstanceHandler) NewContext() TaskContext {
	return &velez_api.CreateHaproxyInstanceTaskPayload{}
}

func (h *createHaproxyInstanceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.CreateHaproxyInstanceTaskPayload)
	if !ok {
		panic("create_haproxy_instance: BuildJobs called with mismatched TaskContext type")
	}

	request := payload.GetRequest()
	serviceName := domain.HaproxyServiceName(request.GetName())

	return []NamedJob{
		{
			Name: stepValidateHaproxyMaster,
			Job: &validateHaproxyMasterJob{
				dataStorage: h.dataStorage,
				runtimes:    h.runtimes,
				environment: request.GetEnvironment(),
				master:      request.GetMasterPgInstance(),
				target:      request.GetConnectionTarget(),
			},
		},
		{
			Name: stepRecordHaproxyDeployBaseline,
			Job: &recordHaproxyDeployBaselineJob{
				jobsEngine: h.jobsEngine,
				payload:    payload,
				entityId:   SmerdEntityID(request.GetEnvironment(), serviceName),
			},
		},
		{
			Name: stepDeployHaproxy,
			Job: &deployHaproxyJob{
				boxes:          h.dataStorage.ResourceBoxes(),
				dataStorage:    h.dataStorage,
				vervServices:   h.vervServices,
				configResolver: h.configResolver,
				request:        request,
				runtimes:       h.runtimes,
			},
		},
		{
			Name: stepWaitHaproxy,
			Job: &waitHaproxyJob{
				payload:     payload,
				jobsEngine:  h.jobsEngine,
				environment: request.GetEnvironment(),
				serviceName: serviceName,
			},
		},
		{
			Name: stepAwaitHaproxyStable,
			Job: &awaitHaproxyStableJob{
				runtimes:    h.runtimes,
				environment: request.GetEnvironment(),
				serviceName: serviceName,
			},
		},
		{
			Name: stepSyncAddresses,
			Job:  &syncAddressesJob{addressBook: h.addressBook, root: staticRootService(serviceName)},
		},
	}
}

type validateHaproxyMasterJob struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver
	environment string
	master      string
	target      *velez_api.HaproxyConnectionTarget
}

func (j *validateHaproxyMasterJob) Do(ctx context.Context) error {
	if !haproxyMasterNamePattern.MatchString(j.master) {
		return rerrors.Wrap(user_errors.ErrHaproxyInvalidMaster)
	}

	svc, err := j.dataStorage.Services().GetByName(ctx, j.master)
	if err != nil {
		return rerrors.Wrap(err, "error getting PostgreSQL instance service")
	}

	_, err = j.dataStorage.PgInstances().GetPgInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting PostgreSQL instance")
	}

	address, err := haproxyEndpointAddress(j.target)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if address != "" {
		return nil
	}

	// The backend name is a Docker DNS alias scoped to the environment network.
	runtime, err := j.runtimes.Runtime(ctx, j.environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy container runtime")
	}

	req := &velez_api.ListSmerds_Request{
		Label: map[string]string{labels.VervServiceLabel: j.master},
	}

	containers, err := runtime.ListContainers(ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error listing PostgreSQL containers")
	}

	if len(containers) == 0 {
		return rerrors.Wrap(user_errors.ErrHaproxyMasterEnvironment)
	}

	return nil
}

// haproxyMasterUsername reads the master pg instance's login role. haproxy:lts
// refuses to start on `option pgsql-check` without `user <username>`, and the
// named role must exist on the backend for the check to authenticate.
func haproxyMasterUsername(ctx context.Context, dataStorage storage.Storage, master string) (string, error) {
	svc, err := dataStorage.Services().GetByName(ctx, master)
	if err != nil {
		return "", rerrors.Wrap(err, "error getting PostgreSQL instance service")
	}

	pg, err := dataStorage.PgInstances().GetPgInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return "", rerrors.Wrap(err, "error getting PostgreSQL instance")
	}

	return pg.Username, nil
}

type deployHaproxyJob struct {
	boxes          vervonomicon.BoxLookup
	dataStorage    storage.Storage
	vervServices   service.VervServicesService
	configResolver service.ServiceConfigResolver
	request        *velez_api.CreateHaproxyInstance_Request
	runtimes       container_runtime.RuntimeResolver
}

func (j *deployHaproxyJob) Do(ctx context.Context) error {
	port := uint32(domain.HaproxyDefaultPort)
	if j.request.Port != nil {
		port = j.request.GetPort()
	}

	user, err := haproxyMasterUsername(ctx, j.dataStorage, j.request.GetMasterPgInstance())
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy master username")
	}

	master := j.request.GetMasterPgInstance()
	target := j.request.GetConnectionTarget()

	address, err := haproxyEndpointAddress(target)
	if err != nil {
		return rerrors.Wrap(err)
	}

	networkName := target.GetDockerNetwork()
	if networkName != "" {
		runtime, runtimeErr := j.runtimes.Runtime(ctx, j.request.GetEnvironment())
		if runtimeErr != nil {
			return rerrors.Wrap(runtimeErr, "error resolving HAProxy container runtime")
		}

		shared, targetErr := resolveHaproxyDockerTarget(ctx, runtime, master, networkName)
		if targetErr != nil {
			return rerrors.Wrap(targetErr)
		}

		address = net.JoinHostPort(shared.host, "5432")
	}

	config, err := renderHaproxyConfigForTarget(port, master, user, address)
	if err != nil {
		return rerrors.Wrap(err)
	}

	descriptor, err := haproxyDescriptor(j.request, port)
	if err != nil {
		return rerrors.Wrap(err)
	}

	resolver := vervonomicon.NewBoxResolver(j.boxes)

	smerdRequest, err := resolver.ResolveRequest(ctx, descriptor, j.request.GetEnvironment(), "")
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy deploy request")
	}

	name := j.request.GetName()
	serviceName := domain.HaproxyServiceName(name)

	smerdRequest.Name = serviceName
	smerdRequest.Labels[labels.VervServiceLabel] = serviceName
	smerdRequest.Labels[labels.DisplayNameLabel] = name
	smerdRequest.Labels[labels.HaproxyInstanceLabel] = name

	// CreateSmerd joins the environment network implicitly only for published ports.
	if len(smerdRequest.GetSettings().GetPorts()) == 0 {
		bind := &velez_api.NetworkBind{NetworkName: env.VervNetwork, Aliases: []string{serviceName}}

		smerdRequest.Settings.Network = append(smerdRequest.Settings.Network, bind)
	}

	if networkName != "" && networkName != env.VervNetwork {
		bind := &velez_api.NetworkBind{NetworkName: networkName}

		smerdRequest.Settings.Network = append(smerdRequest.Settings.Network, bind)
	}

	err = j.configResolver.WriteFile(ctx, smerdRequest, domain.HaproxyConfigPath, config)
	if err != nil {
		return rerrors.Wrap(err, "error writing HAProxy config")
	}

	deployReq := domain.CreateDeployReq{
		ServiceName:    serviceName,
		DisplayName:    name,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating HAProxy deploy")
	}

	return nil
}

func haproxyDescriptor(request *velez_api.CreateHaproxyInstance_Request, port uint32) (verv.Descriptor, error) {
	files, err := builtin.Read(haproxyDescriptorName)
	if err != nil {
		return verv.Descriptor{}, rerrors.Wrap(err, "error reading HAProxy descriptor")
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, request.GetEnvironment())
	if err != nil {
		return verv.Descriptor{}, rerrors.Wrap(err, "error merging HAProxy environment overlay")
	}

	descriptor.Source = verv.SourceKindBuiltin
	if request.GetBox() != "" {
		descriptor.Deployment.App.Box = request.GetBox()
	}

	if request.PublishPort != nil && !request.GetPublishPort() {
		descriptor.Deployment.App.Ports = nil
	} else {
		for i := range descriptor.Deployment.App.Ports {
			mapping := &descriptor.Deployment.App.Ports[i]

			mapping.Port = int(port)
			mapping.ExposeTo = int(port)
		}
	}

	volumeNames := map[string]string{haproxyDescriptorVolume: domain.HaproxyConfigVolumeName(request.GetName())}
	renameDescriptorVolumes(&descriptor, volumeNames)

	return descriptor, nil
}

type recordHaproxyDeployBaselineJob struct {
	jobsEngine taskWatcher
	payload    *velez_api.CreateHaproxyInstanceTaskPayload
	entityId   string
}

func (j *recordHaproxyDeployBaselineJob) Do(ctx context.Context) error {
	return recordDeployBaseline(ctx, j.jobsEngine, j.payload, j.entityId)
}

type waitHaproxyJob struct {
	payload     *velez_api.CreateHaproxyInstanceTaskPayload
	jobsEngine  taskWatcher
	environment string
	serviceName string
}

func (j *waitHaproxyJob) Do(ctx context.Context) error {
	return waitForSmerdDeployAfter(
		ctx, j.jobsEngine, j.environment, j.serviceName, j.payload.GetDeployBaselineTaskId(),
	)
}

func renderHaproxyConfig(port uint32, master, user string) ([]byte, error) {
	return renderHaproxyConfigForTarget(port, master, user, "")
}

func renderHaproxyConfigForTarget(port uint32, master, user, address string) ([]byte, error) {
	if port == 0 || port > domain.HaproxyMaxPort {
		return nil, rerrors.Wrap(user_errors.ErrHaproxyInvalidPort)
	}

	if !haproxyMasterNamePattern.MatchString(master) {
		return nil, rerrors.Wrap(user_errors.ErrHaproxyInvalidMaster)
	}

	if !haproxyMasterNamePattern.MatchString(user) {
		return nil, rerrors.New("HAProxy master pg instance has an unusable username: " + user)
	}

	if address == "" {
		address = net.JoinHostPort(master, "5432")
	}

	params := struct {
		Port    uint32
		Master  string
		User    string
		Address string
	}{Port: port, Master: master, User: user, Address: address}

	var config bytes.Buffer

	err := haproxyTemplate.Execute(&config, params)
	if err != nil {
		return nil, rerrors.Wrap(err, "error rendering HAProxy config")
	}

	return config.Bytes(), nil
}
