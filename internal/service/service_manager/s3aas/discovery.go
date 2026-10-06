package s3aas

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/garage"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	tcpProtocol = "tcp"
	httpScheme  = "http://"
	hostName    = "localhost"

	defaultReplicationFactor = 1
)

type instanceRef struct {
	name        string
	environment string
	remoteHost  string
	runtime     container_runtime.ContainerRuntime
	instance    container.Summary
	webUi       *container.Summary
}

func (s *Service) discoverInstances(ctx context.Context) ([]instanceRef, error) {
	environments, err := s.dataStorage.Environments().ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments")
	}

	refs := make([]instanceRef, 0)

	for _, environment := range environments {
		found, discoverErr := s.discoverInEnvironment(ctx, environment)
		if discoverErr != nil {
			return nil, rerrors.Wrap(discoverErr)
		}

		refs = append(refs, found...)
	}

	return refs, nil
}

func (s *Service) discoverInEnvironment(
	ctx context.Context,
	environment domain.Environment,
) ([]instanceRef, error) {
	runtime, err := s.runtimes.Runtime(ctx, environment.Name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	listReq := &velez_api.ListSmerds_Request{Environment: environment.Name}

	list, err := runtime.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	webUis := make(map[string]container.Summary)

	for _, summary := range list {
		name, isWebUi := summary.Labels[labels.S3WebUiLabel]
		if isWebUi {
			webUis[name] = summary
		}
	}

	refs := make([]instanceRef, 0)

	for _, summary := range list {
		name, isInstance := summary.Labels[labels.S3InstanceLabel]
		if !isInstance {
			continue
		}

		ref := instanceRef{
			name:        name,
			environment: environment.Name,
			remoteHost:  environment.RemoteHost(s.docker.Host()),
			runtime:     runtime,
			instance:    summary,
		}

		webUi, hasWebUi := webUis[name]
		if hasWebUi {
			ref.webUi = &webUi
		}

		refs = append(refs, ref)
	}

	return refs, nil
}

func (s *Service) findInstance(ctx context.Context, name string) (instanceRef, error) {
	refs, err := s.discoverInstances(ctx)
	if err != nil {
		return instanceRef{}, rerrors.Wrap(err)
	}

	for _, ref := range refs {
		if ref.name == name {
			return ref, nil
		}
	}

	return instanceRef{}, rerrors.Wrap(user_errors.ErrS3InstanceNotFound)
}

func (s *Service) connect(ctx context.Context, name string) (instanceRef, *garage.Client, error) {
	ref, err := s.findInstance(ctx, name)
	if err != nil {
		return instanceRef{}, nil, rerrors.Wrap(err)
	}

	adminToken, err := s.secretsStore.Get(ctx, domain.S3AdminTokenSecretRef(name))
	if err != nil {
		return instanceRef{}, nil, rerrors.Wrap(err, "error getting garage admin token")
	}

	adminUrl, err := garage.AdminUrl(ctx, ref.runtime, domain.S3ServiceName(name))
	if err != nil {
		return instanceRef{}, nil, rerrors.Wrap(err, "error resolving garage admin url")
	}

	return ref, garage.New(adminUrl, adminToken), nil
}

func (s *Service) describe(ctx context.Context, ref instanceRef) domain.S3Instance {
	instance := domain.S3Instance{
		Name:              ref.name,
		S3Port:            hostPort(ref.instance, domain.S3ApiContainerPort),
		Region:            domain.S3DefaultRegion,
		ReplicationFactor: defaultReplicationFactor,
		Environment:       ref.environment,
		RemoteHost:        ref.remoteHost,
		Status:            ref.instance.State,
		CreatedAt:         time.Unix(ref.instance.Created, 0),
	}

	if ref.webUi != nil {
		instance.WebUiPort = hostPort(*ref.webUi, domain.S3WebUiContainerPort)
	}

	content, err := s.configResolver.ReadFile(ctx, domain.S3ServiceName(ref.name), ref.environment, domain.S3ConfigPath)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("instance", ref.name).
			Err(err).
			Msg("error reading s3 instance config, falling back to defaults")

		return instance
	}

	region, replicationFactor := parseGarageConfig(content)

	if region != "" {
		instance.Region = region
	}

	if replicationFactor != 0 {
		instance.ReplicationFactor = replicationFactor
	}

	return instance
}

func hostPort(summary container.Summary, privatePort uint16) uint32 {
	for _, port := range summary.Ports {
		if port.PrivatePort == privatePort && port.Type == tcpProtocol && port.PublicPort != 0 {
			return uint32(port.PublicPort)
		}
	}

	return 0
}

// s3Endpoint takes the prefixed service name. It mirrors the jobs package's
// registryInstanceUrl: a Velez inside a container reaches the instance over
// the docker network, a bare binary only through the host-published port.
// An instance on a remote daemon is reachable by neither, only through the
// remote host's published port.
func s3Endpoint(serviceName, remoteHost string, publishedPort uint32) string {
	if remoteHost != "" {
		return publishedEndpoint(remoteHost, publishedPort)
	}

	if env.IsInContainer() {
		return internalEndpoint(serviceName, domain.S3ApiContainerPort)
	}

	return publishedEndpoint(hostName, publishedPort)
}

func publishedEndpoint(host string, publishedPort uint32) string {
	return httpScheme + net.JoinHostPort(host, strconv.Itoa(int(publishedPort)))
}

func internalEndpoint(host string, port int) string {
	return httpScheme + net.JoinHostPort(host, strconv.Itoa(port))
}

func webUiEndpoint(webUiServiceName, remoteHost string, publishedPort uint32) string {
	if remoteHost != "" {
		return publishedEndpoint(remoteHost, publishedPort)
	}

	if env.IsInContainer() {
		return internalEndpoint(webUiServiceName, domain.S3WebUiContainerPort)
	}

	return publishedEndpoint(hostName, publishedPort)
}
