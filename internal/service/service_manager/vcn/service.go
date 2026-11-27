package vcn

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	providerHeadscale = "headscale"
	sourceExternal    = "external"
	sourceDeployed    = "deployed"
)

type Service struct {
	storage  *storage.Container
	runtimes container_runtime.RuntimeResolver
	node     node_clients.NodeClients
	engine   jobs.Engine
}

func New(
	dataStorage *storage.Container, runtimes container_runtime.RuntimeResolver,
	node node_clients.NodeClients, engine jobs.Engine,
) *Service {
	return &Service{storage: dataStorage, runtimes: runtimes, node: node, engine: engine}
}

func (s *Service) Create(ctx context.Context, req service.CreateVcnNetworkRequest) (domain.Network, error) {
	if req.External != nil && !s.storage.IsStatefull() {
		return domain.Network{}, rerrors.Wrap(user_errors.ErrExternalNetworkRequiresCluster)
	}

	err := domain.ValidateInstanceName(req.Name)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	record := domain.Network{
		Id:        uuid.NewString(),
		Name:      req.Name,
		Provider:  providerHeadscale,
		CreatedAt: time.Now().UTC(),
	}

	switch {
	case req.External != nil:
		return s.createExternal(ctx, record, *req.External)
	case req.Deploy != nil:
		return s.createDeployed(ctx, record, req.Host, *req.Deploy)
	default:
		return domain.Network{}, rerrors.Wrap(user_errors.ErrVcnNetworkSourceRequired)
	}
}

func (s *Service) List(ctx context.Context) ([]domain.Network, error) {
	repo := s.storage.Networks()
	if repo == nil {
		return s.listLocal(ctx, "")
	}

	records, err := repo.List(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return records, nil
}

func (s *Service) Get(ctx context.Context, id string) (service.VcnNetworkDetails, error) {
	record, err := s.find(ctx, id)
	if err != nil {
		return service.VcnNetworkDetails{}, rerrors.Wrap(err)
	}

	backend, err := s.backend(ctx, record)
	if err != nil {
		return service.VcnNetworkDetails{}, rerrors.Wrap(err)
	}

	peers, err := backend.ListNodes(ctx)
	if err != nil {
		return service.VcnNetworkDetails{}, rerrors.Wrap(err, "error listing VPN peers")
	}

	keys, err := backend.ListApiKeys(ctx)
	if err != nil {
		return service.VcnNetworkDetails{}, rerrors.Wrap(err, "error listing VPN API keys")
	}

	return service.VcnNetworkDetails{
		Network: record,
		Peers:   peers,
		ApiKeys: keys,
	}, nil
}

func (s *Service) CurrentNode(ctx context.Context) (string, error) {
	repo := s.storage.Networks()
	if repo == nil {
		return "", nil
	}

	name, err := repo.CurrentNode(ctx)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	return name, nil
}

func (s *Service) find(ctx context.Context, id string) (domain.Network, error) {
	if id == "" {
		return domain.Network{}, rerrors.Wrap(user_errors.ErrVcnNetworkNotFound)
	}

	repo := s.storage.Networks()
	if repo != nil {
		record, err := repo.Get(ctx, id)
		if err != nil {
			if rerrors.Is(err, user_errors.ErrStorageNotFound) {
				return domain.Network{}, rerrors.Wrap(user_errors.ErrVcnNetworkNotFound)
			}

			return domain.Network{}, rerrors.Wrap(err)
		}

		return record, nil
	}

	networks, err := s.listLocal(ctx, id)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	if len(networks) == 0 {
		return domain.Network{}, rerrors.Wrap(user_errors.ErrVcnNetworkNotFound)
	}

	return networks[0], nil
}

func (s *Service) createExternal(
	ctx context.Context, record domain.Network, external service.VcnExternalSource,
) (domain.Network, error) {
	repo := s.storage.Networks()
	if repo == nil {
		return domain.Network{}, rerrors.Wrap(user_errors.ErrExternalNetworkRequiresCluster)
	}

	base, err := url.Parse(strings.TrimSpace(external.Url))
	if err != nil || base.Hostname() == "" || base.User != nil || base.RawQuery != "" || base.ForceQuery ||
		base.Fragment != "" || (base.Scheme != "http" && base.Scheme != "https") {
		return domain.Network{}, rerrors.Wrap(user_errors.ErrVcnInvalidHeadscaleUrl)
	}

	token := strings.TrimSpace(external.Token)
	if token == "" {
		return domain.Network{}, rerrors.Wrap(user_errors.ErrVcnHeadscaleTokenRequired)
	}

	record.Source = sourceExternal
	record.Url = strings.TrimRight(base.String(), "/")
	record.Token = token

	_, err = headscale.Connect(ctx, record.Url, record.Token)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err, "error verifying Headscale URL and API token")
	}

	persisted, err := persist(ctx, repo, record)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(err)
	}

	persisted.Token = ""

	return persisted, nil
}

func persist(ctx context.Context, repo storage.NetworksStorage, record domain.Network) (domain.Network, error) {
	err := repo.Create(ctx, record)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageAlreadyExists) {
			return domain.Network{}, rerrors.Wrap(user_errors.ErrVcnNetworkExists)
		}

		return domain.Network{}, rerrors.Wrap(err)
	}

	return record, nil
}
