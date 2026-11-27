package vcn

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	headscalePattern "go.vervstack.ru/Velez/internal/patterns/headscale"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	networkSourceLabel = "velez.vcn.source"
	networkIdLabel     = "velez.vcn.id"
	networkNameLabel   = "velez.vcn.name"
	networkUrlLabel    = "velez.vcn.url"

	serviceNamePrefix = "vcn-"
	configFile        = "/etc/headscale/config.yaml"
	apiKeyFile        = "/etc/headscale/.velez-api-key"
	headscalePort     = 8080
	maxPort           = 65535
	configMode        = 0o644
	keyMode           = 0o600

	startupTimeout = 30 * time.Second
	retryInterval  = 500 * time.Millisecond
	cleanupTimeout = 15 * time.Second
)

func (s *Service) createDeployed(
	ctx context.Context, record domain.Network, requestHost string, request domain.SetupHeadscaleRequest,
) (domain.Network, error) {
	runtime, err := s.runtimes.Runtime(ctx, "")
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	record.Source = sourceDeployed
	record.ServiceName = serviceNamePrefix + record.Name

	repo := s.storage.Networks()
	if repo != nil {
		record.NodeName, err = repo.CurrentNode(ctx)
		if err != nil {
			return domain.Network{}, rerrors.Wrap(err, "error resolving deployment node")
		}
	}

	host, err := s.loginHost(ctx, requestHost, repo != nil)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	port, isReserved, err := s.reservePort(request.ExposeToPort)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	created := false

	defer func() {
		if isReserved && !created {
			s.node.PortManager().UnlockPorts([]uint32{port})
		}
	}()

	portText := strconv.FormatUint(uint64(port), 10)

	request.ExposeToPort = &portText
	record.Url = "http://" + net.JoinHostPort(host, portText)

	releaseContainer, err := s.startHeadscale(ctx, runtime, &record, request)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	defer func() {
		if !created {
			releaseContainer()
		}
	}()

	if repo != nil {
		_, err = persist(ctx, repo, record)
		if err != nil {
			return domain.Network{}, rerrors.Wrap(err)
		}

		err = s.storage.Services().UpsertService(ctx, record.ServiceName, record.Name)
		if err != nil {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanupTimeout)
			defer cancel()

			logCleanupError(repo.Drop(cleanupCtx, record.Id), record.Id)

			return domain.Network{}, rerrors.Wrap(err, "error registering Headscale service")
		}
	}

	created = true
	record.Token = ""

	return record, nil
}

func (s *Service) startHeadscale(
	ctx context.Context, runtime container_runtime.ContainerRuntime,
	record *domain.Network, request domain.SetupHeadscaleRequest,
) (func(), error) {
	spec := deployedSpec(*record, request)

	_, err := runtime.PullImage(ctx, spec.Config.Image)
	if err != nil {
		return nil, rerrors.Wrap(err, "error pulling Headscale image")
	}

	created, err := runtime.ContainerCreate(ctx, spec)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating Headscale container")
	}

	record.ContainerId = created.ID

	remove := func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanupTimeout)
		defer cancel()

		removeErr := runtime.Remove(cleanupCtx, created.ID)
		logCleanupError(removeErr, record.Id)

		if removeErr == nil {
			logCleanupError(runtime.RemoveVolume(cleanupCtx, deployedVolume(*record)), record.Id)
		}
	}

	err = s.configureHeadscale(ctx, runtime, record)
	if err != nil {
		remove()

		return nil, rerrors.Wrap(err)
	}

	return remove, nil
}

func (s *Service) configureHeadscale(
	ctx context.Context, runtime container_runtime.ContainerRuntime, record *domain.Network,
) error {
	config := headscalePattern.ConfigForNetwork(record.Url)

	err := runtime.CopyToContainer(ctx, record.ContainerId, configFile, config, configMode)
	if err != nil {
		return rerrors.Wrap(err, "error writing Headscale configuration")
	}

	start := container.StartOptions{}

	err = s.node.Docker().Client().ContainerStart(ctx, record.ContainerId, start)
	if err != nil {
		return rerrors.Wrap(err, "error starting Headscale container")
	}

	err = s.waitForHeadscale(ctx, runtime, record.ContainerId)
	if err != nil {
		return rerrors.Wrap(err, "Headscale failed to become ready")
	}

	record.Token, err = headscale.IssueContainerApiKey(ctx, s.node.Docker(), record.ContainerId)
	if err != nil {
		return rerrors.Wrap(err, "error creating Headscale management API key")
	}

	err = runtime.CopyToContainer(ctx, record.ContainerId, apiKeyFile, []byte(record.Token), keyMode)
	if err != nil {
		return rerrors.Wrap(err, "error storing Headscale management API key")
	}

	return nil
}

func (s *Service) waitForHeadscale(ctx context.Context, runtime container_runtime.ContainerRuntime, id string) error {
	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	for {
		info, found, err := runtime.Inspect(ctx, id)
		if err != nil {
			return rerrors.Wrap(err)
		}

		if !found || info.State == nil || !info.State.Running {
			return rerrors.Wrap(user_errors.ErrVcnHeadscaleNotRunning)
		}

		if info.State.Health != nil && info.State.Health.Status == container.Healthy {
			return nil
		}

		select {
		case <-ctx.Done():
			return rerrors.Wrap(ctx.Err())
		case <-ticker.C:
		}
	}
}

func (s *Service) reservePort(custom *string) (port uint32, isReserved bool, err error) {
	if custom != nil {
		parsed, parseErr := strconv.ParseUint(*custom, 10, 16)
		if parseErr != nil || parsed == 0 || parsed > maxPort {
			return 0, false, rerrors.Wrap(user_errors.ErrVcnInvalidPort)
		}

		return uint32(parsed), false, nil
	}

	port, err = s.node.PortManager().GetPort()
	if err != nil {
		return 0, false, rerrors.Wrap(err, "error selecting a published Headscale port")
	}

	return port, true, nil
}

func (s *Service) loginHost(ctx context.Context, requestHost string, clusterMode bool) (string, error) {
	if requestHost != "" && !strings.EqualFold(requestHost, "localhost") && !net.ParseIP(requestHost).IsLoopback() {
		return requestHost, nil
	}

	if clusterMode {
		return "", rerrors.Wrap(user_errors.ErrVcnReachableHostRequired)
	}

	options := network.InspectOptions{}

	bridge, err := s.node.Docker().Client().NetworkInspect(ctx, "bridge", options)
	if err != nil {
		return "", rerrors.Wrap(err, "error resolving Docker host address")
	}

	for _, config := range bridge.IPAM.Config {
		if config.Gateway != "" {
			return config.Gateway, nil
		}
	}

	return "", rerrors.Wrap(user_errors.ErrVcnReachableHostRequired)
}

func (s *Service) listLocal(ctx context.Context, id string) ([]domain.Network, error) {
	runtime, err := s.runtimes.Runtime(ctx, "")
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	req := &pb.ListSmerds_Request{Label: map[string]string{networkSourceLabel: sourceDeployed}}
	if id != "" {
		req.Label[networkIdLabel] = id
	}

	containers, err := runtime.ListContainers(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing VPN network containers")
	}

	records := make([]domain.Network, 0, len(containers))

	for _, c := range containers {
		if c.Labels[networkIdLabel] == "" {
			continue
		}

		record := domain.Network{
			Id:          c.Labels[networkIdLabel],
			Name:        c.Labels[networkNameLabel],
			Provider:    providerHeadscale,
			Source:      sourceDeployed,
			Url:         c.Labels[networkUrlLabel],
			ServiceName: c.Labels[labels.VervServiceLabel],
			ContainerId: c.ID,
			CreatedAt:   time.Unix(c.Created, 0).UTC(),
		}

		records = append(records, record)
	}

	return records, nil
}

func (s *Service) withCredentials(ctx context.Context, record domain.Network) (domain.Network, error) {
	if record.Source != sourceDeployed {
		return record, nil
	}

	nodeName, err := s.CurrentNode(ctx)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	if nodeName != record.NodeName {
		return record, nil
	}

	runtime, err := s.runtimes.Runtime(ctx, "")
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	info, found, err := runtime.Inspect(ctx, record.ContainerId)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	if !found || info.Config == nil || info.Config.Labels[networkIdLabel] != record.Id {
		return domain.Network{}, rerrors.Wrap(user_errors.ErrVcnHeadscaleNotRunning)
	}

	if record.Token == "" {
		key, readErr := runtime.CopyFromContainer(ctx, record.ContainerId, apiKeyFile)
		if readErr != nil {
			return domain.Network{}, rerrors.Wrap(readErr, "error reading Headscale management API key")
		}

		record.Token = strings.TrimSpace(string(key))
	}

	address, err := runtime.ContainerAddress(info, headscalePort)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	endpoint := url.URL{Scheme: "http", Host: address}

	record.Url = endpoint.String()

	return record, nil
}

func deployedSpec(
	record domain.Network, request domain.SetupHeadscaleRequest,
) container_runtime.ContainerCreateRequest {
	spec := headscalePattern.Headscale(request)

	spec.Hostname = record.ServiceName
	spec.Labels[labels.VervServiceLabel] = record.ServiceName
	spec.Labels[labels.DisplayNameLabel] = record.Name
	spec.Labels[networkSourceLabel] = sourceDeployed
	spec.Labels[networkIdLabel] = record.Id
	spec.Labels[networkNameLabel] = record.Name
	spec.Labels[networkUrlLabel] = record.Url
	spec.HostConfig.RestartPolicy.Name = container.RestartPolicyUnlessStopped
	spec.Healthcheck.Interval = retryInterval

	// Each named network has its own SQLite database and private keys.
	for index := range spec.HostConfig.Mounts {
		spec.HostConfig.Mounts[index].Source = deployedVolume(record)
	}

	port := nat.Port(headscalePattern.APIPort + "/tcp")

	spec.ExposedPorts[port] = spec.ExposedPorts[headscalePattern.APIPort]
	spec.HostConfig.PortBindings[port] = spec.HostConfig.PortBindings[headscalePattern.APIPort]
	delete(spec.ExposedPorts, headscalePattern.APIPort)
	delete(spec.HostConfig.PortBindings, headscalePattern.APIPort)

	config := &container_runtime.ContainerConfig{Config: spec.Config}
	hostConfig := &container_runtime.HostConfig{HostConfig: spec.HostConfig}

	return container_runtime.ContainerCreateRequest{
		Config:        config,
		HostConfig:    hostConfig,
		ContainerName: record.ServiceName,
	}
}

func deployedVolume(record domain.Network) string {
	return serviceNamePrefix + record.Id
}

func logCleanupError(err error, id string) {
	if err != nil {
		log.Error().Err(err).Str("network_id", id).Msg("error cleaning up VPN network")
	}
}
