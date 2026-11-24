package jobs

import (
	"bytes"
	"context"
	"regexp"
	"text/template"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
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
    server {{.Master}} {{.Master}}:5432 check
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

	config, err := renderHaproxyConfig(port, j.request.GetMasterPgInstance(), user)
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

	for i := range descriptor.Deployment.App.Ports {
		mapping := &descriptor.Deployment.App.Ports[i]

		mapping.Port = int(port)
		mapping.ExposeTo = int(port)
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
	if port == 0 || port > domain.HaproxyMaxPort {
		return nil, rerrors.Wrap(user_errors.ErrHaproxyInvalidPort)
	}

	if !haproxyMasterNamePattern.MatchString(master) {
		return nil, rerrors.Wrap(user_errors.ErrHaproxyInvalidMaster)
	}

	if !haproxyMasterNamePattern.MatchString(user) {
		return nil, rerrors.New("HAProxy master pg instance has an unusable username: " + user)
	}

	params := struct {
		Port   uint32
		Master string
		User   string
	}{Port: port, Master: master, User: user}

	var config bytes.Buffer

	err := haproxyTemplate.Execute(&config, params)
	if err != nil {
		return nil, rerrors.Wrap(err, "error rendering HAProxy config")
	}

	return config.Bytes(), nil
}
