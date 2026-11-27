package vcn

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	networkNamespace = "velez"
)

// provider is scoped to one network. Its node, key and registration operations
// have no dependency on where the provider is hosted.
type provider interface {
	ListNodes(ctx context.Context) ([]domain.VcnNode, error)
	ListApiKeys(ctx context.Context) ([]domain.VcnApiKey, error)
	IssueApiKey(ctx context.Context) (string, error)
	ExpireApiKey(ctx context.Context, prefix string) error
	RegisterNode(ctx context.Context, req domain.RegisterVcnNodeReq) error
	GetNamespace(ctx context.Context, name string) (domain.VcnNamespace, error)
	CreateNamespace(ctx context.Context, name string) (domain.VcnNamespace, error)
	GetClientAuthKey(ctx context.Context, req domain.GetVcnAuthKeyReq) (domain.VcnAuthKey, error)
	IssueClientKey(ctx context.Context, req domain.IssueClientKey) (string, error)
}

func (s *Service) Peers(ctx context.Context, id string) ([]domain.VcnNode, error) {
	p, err := s.resolveBackend(ctx, id)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	nodes, err := p.ListNodes(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return nodes, nil
}

func (s *Service) IssueApiKey(ctx context.Context, id string) (string, error) {
	p, err := s.resolveBackend(ctx, id)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	key, err := p.IssueApiKey(ctx)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	return key, nil
}

func (s *Service) ExpireApiKey(ctx context.Context, id, prefix string) error {
	if prefix == "" {
		return rerrors.Wrap(user_errors.ErrVcnKeyPrefixRequired)
	}

	record, err := s.find(ctx, id)
	if err != nil {
		return rerrors.Wrap(err)
	}

	record, err = s.withCredentials(ctx, record)
	if err != nil {
		return rerrors.Wrap(err)
	}

	managementPrefix, _, _ := strings.Cut(record.Token, ".")
	if prefix == managementPrefix {
		return rerrors.Wrap(user_errors.ErrVcnManagementKeyProtected)
	}

	p, err := newProvider(record)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = p.ExpireApiKey(ctx, prefix)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (s *Service) Connection(ctx context.Context, id string) (loginUrl, authKey string, err error) {
	record, err := s.find(ctx, id)
	if err != nil {
		return "", "", rerrors.Wrap(err)
	}

	p, err := s.backend(ctx, record)
	if err != nil {
		return "", "", rerrors.Wrap(err)
	}

	namespace, err := ensureNamespace(ctx, p)
	if err != nil {
		return "", "", rerrors.Wrap(err, "error preparing VPN namespace")
	}

	lookup := domain.GetVcnAuthKeyReq{NamespaceId: namespace.Id, ReusableOnly: true}

	key, err := p.GetClientAuthKey(ctx, lookup)
	if err != nil && !rerrors.Is(err, user_errors.ErrNotFound) {
		return "", "", rerrors.Wrap(err, "error reading VPN auth key")
	}

	if key.Key != "" {
		return record.Url, key.Key, nil
	}

	issue := domain.IssueClientKey{NamespaceId: namespace.Id, Reusable: true}

	authKey, err = p.IssueClientKey(ctx, issue)
	if err != nil {
		return "", "", rerrors.Wrap(err, "error issuing VPN auth key")
	}

	return record.Url, authKey, nil
}

func ensureNamespace(ctx context.Context, p provider) (domain.VcnNamespace, error) {
	namespace, err := p.GetNamespace(ctx, networkNamespace)
	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err)
	}

	if namespace.Id != "" {
		return namespace, nil
	}

	namespace, err = p.CreateNamespace(ctx, networkNamespace)
	if rerrors.Is(err, user_errors.ErrHeadscaleNamespaceAlreadyExists) {
		namespace, err = p.GetNamespace(ctx, networkNamespace)
	}

	if err != nil {
		return domain.VcnNamespace{}, rerrors.Wrap(err)
	}

	if namespace.Id == "" {
		return domain.VcnNamespace{}, rerrors.Wrap(user_errors.ErrVcnNamespaceIdMissing)
	}

	return namespace, nil
}

func (s *Service) resolveBackend(ctx context.Context, id string) (provider, error) {
	record, err := s.find(ctx, id)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return s.backend(ctx, record)
}

func (s *Service) backend(ctx context.Context, record domain.Network) (provider, error) {
	record, err := s.withCredentials(ctx, record)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return newProvider(record)
}

func newProvider(record domain.Network) (provider, error) {
	if record.Provider != providerHeadscale {
		return nil, rerrors.Wrap(user_errors.ErrVcnProviderUnsupported)
	}

	return headscale.New(record.Url, record.Token), nil
}
