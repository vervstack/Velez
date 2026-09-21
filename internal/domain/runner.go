package domain

import (
	"time"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// Runner is the runner-specific satellite row for a Runners-as-a-Service
// instance (velez.runners). Status and environment live on the underlying
// velez.services / deployment_specifications / deployments rows and are read
// through VervServicesService and ListSmerds - never duplicated here. Mirrors
// PgInstance / PostgresAPI's division of responsibility, generalized to be
// git-provider-agnostic.
type Runner struct {
	ServiceID int64
	Provider  string
	Scope     string
	Target    string
	Labels    []string
	// SecretRef - the canonical "scope/owner/key" string form of the
	// domain.SecretRef the access token is stored under (see
	// SecretRef.String). Never the value itself - resolved only through
	// internal/service/secrets.Store.
	SecretRef string
	// BaseUrl - the git provider instance's base URL (e.g. a self-managed
	// GitLab's URL). Empty for a provider with a single fixed API host
	// (GitHub).
	BaseUrl   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	runnerSecretScope = "runneraas"
	runnerSecretKey   = "access_token"
	// runnerTokenSecretKey stores the short-lived registration token minted
	// at CreateRunner time - GetRunnerCredentials resolves it back out for a
	// provider whose registration step isn't automated (GitLab in v1). A
	// self-registering provider (GitHub) never reads it back; it's stored
	// for every provider regardless, so GetRunnerCredentials stays
	// provider-agnostic rather than branching on whether it exists.
	runnerTokenSecretKey = "registration_token"
)

// RunnerAccessTokenSecretRef is the secret ref a runner's caller-supplied
// access token is stored under. Shared between runneraas (CreateRunner's
// caller-facing validation) and internal/jobs/create_runner.go (which does
// the actual Put) - both import domain, avoiding an import cycle between
// them. Mirrors DockerSocketGrantSecretRef's role as the one place a
// well-known secret ref is derived.
func RunnerAccessTokenSecretRef(name string) SecretRef {
	return SecretRef{Scope: runnerSecretScope, Owner: name, Key: runnerSecretKey}
}

// RunnerRegistrationTokenSecretRef is the secret ref a runner's minted/
// pass-through registration token is stored under. See
// RunnerAccessTokenSecretRef's doc comment.
func RunnerRegistrationTokenSecretRef(name string) SecretRef {
	return SecretRef{Scope: runnerSecretScope, Owner: name, Key: runnerTokenSecretKey}
}

// UpsertRunnerReq creates or replaces the velez.runners row for a service.
type UpsertRunnerReq struct {
	ServiceID int64
	Provider  string
	Scope     string
	Target    string
	Labels    []string
	SecretRef string
	BaseUrl   string
}

// CreateRunnerReq is the input to RunnersService.CreateRunner. See
// runners_api.proto's CreateRunner.Request. Provider and AccessToken are
// resolved by the transport layer from the request's provider_config oneof
// before this struct is built.
type CreateRunnerReq struct {
	Name     string
	Provider velez_api.RunnerProvider
	Scope    velez_api.RunnerScope
	Target   string
	Labels   []string

	// AccessToken - a token with permission to create a runner registration
	// token for Target (a GitHub PAT or a GitLab PAT, depending on Provider).
	// Never stored as given; see internal/service/secrets.
	AccessToken string

	// BaseUrl - the git provider instance's base URL. Only meaningful for a
	// provider with more than one possible host (GitLab); empty for GitHub.
	BaseUrl string

	// Environment - the isolated namespace this runner deploys into. Empty
	// means the default/PROD environment (see storage/environments.Resolve).
	Environment string

	// DockerSocketAddress - tcp:// address of a Docker daemon this runner
	// should talk to instead of the default host socket. Empty means Velez
	// grants the host socket via its existing internal-only bind-mount gate.
	// See runners_api.proto's CreateRunner.Request.docker_socket_address.
	DockerSocketAddress string
}

// RunnerView is one resolved Runners-as-a-Service instance: runner-specific
// facts (Runner) merged with live service/deployment state read through
// VervServicesService - status and environment are never persisted alongside
// the runner-specific facts.
type RunnerView struct {
	Name        string
	Provider    velez_api.RunnerProvider
	Scope       velez_api.RunnerScope
	Target      string
	Labels      []string
	BaseUrl     string
	Environment string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListRunnersReq pages through every Runners-as-a-Service instance.
type ListRunnersReq struct {
	Paging Paging
}

// RunnerList is the paged result of RunnersService.ListRunners.
type RunnerList struct {
	Total   uint64
	Runners []RunnerView
}

// RunnerCredentials is the result of RunnersService.GetRunnerCredentials -
// the only RunnersService operation that resolves a secret_ref to its
// plaintext value.
type RunnerCredentials struct {
	Token    string
	Target   string
	Provider velez_api.RunnerProvider
}
