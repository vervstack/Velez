package jobs

import (
	"context"
	"fmt"
	"maps"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	stepVerifyRegistryLogin = "verify_registry_login"
	stepStoreRegistrySecret = "store_registry_secret"
	stepUpsertRegistryRow   = "upsert_registry_row"

	registryPendingSecretScope = "registryaas-pending"

	registryAuthHtpasswd      = "htpasswd"
	registryHttpAddrEnvVar    = "REGISTRY_HTTP_ADDR"
	registryApiPath           = "/v2/"
	registeredRegistryTimeout = 10 * time.Second
)

type registerRegistryAccessor interface {
	GetRegistryUsername() string
	GetRegistryPendingSecretOwner() string
}

type registerRegistrySetter interface {
	SetRegistryUsername(v string)
}

type registryLogin struct {
	username string
	password string
}

// RegisteredRegistryPendingSecretRef holds the password a register request
// supplied until store_registry_secret moves it to its final ref; owner is the
// task entity id.
func RegisteredRegistryPendingSecretRef(owner string) domain.SecretRef {
	return domain.SecretRef{Scope: registryPendingSecretScope, Owner: owner, Key: registryaasSecretKey}
}

func deletePendingRegistrySecret(ctx context.Context, store secrets.Store, owner string) error {
	if owner == "" {
		return nil
	}

	err := store.Delete(ctx, RegisteredRegistryPendingSecretRef(owner))
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting pending registry secret")
	}

	return nil
}

// isRegistryAuthConfigured tells whether the registry container demands a
// login: only htpasswd auth is verifiable, and it is configured through env.
func isRegistryAuthConfigured(containerEnv map[string]string) bool {
	return containerEnv[envRegistryAuth] == registryAuthHtpasswd ||
		containerEnv[envRegistryAuthHtpasswdPath] != ""
}

// hasRegistryLogin is true once verify_registry_login found auth configured:
// it clears the username otherwise.
func hasRegistryLogin(registry registerRegistryAccessor) bool {
	return registry.GetRegistryUsername() != ""
}

// registryContainerPort is the port the registry listens on inside the
// container: REGISTRY_HTTP_ADDR when it names one, the image default otherwise.
func registryContainerPort(containerEnv map[string]string) int32 {
	_, portValue, err := net.SplitHostPort(containerEnv[registryHttpAddrEnvVar])
	if err != nil {
		return registryaasContainerPort
	}

	port, err := strconv.ParseUint(portValue, 10, 32)
	if err != nil || port == 0 {
		return registryaasContainerPort
	}

	return int32(port) //nolint:gosec
}

// registryPortFromMappings is the host port the registry is published on, or
// its container port when nothing publishes it.
func registryPortFromMappings(ports []*velez_api.Port, containerPort int32) int32 {
	for _, port := range ports {
		if port.GetServicePortNumber() == uint32(containerPort) && port.ExposedTo != nil { //nolint:gosec
			return int32(port.GetExposedTo()) //nolint:gosec
		}
	}

	return containerPort
}

func registryInspectedPort(info container.InspectResponse) int32 {
	containerPort := registryContainerPort(parser.ToDockerEnv(info.Config.Env))

	return registryPortFromMappings(parser.ToPortsFromInspect(info), containerPort)
}

// registeredRegistryLabels are the labels the local_storage backend recovers a
// registry instance's facts from - see labels.RegistryaasInstanceLabel.
func registeredRegistryLabels(username string, port int32) map[string]string {
	return map[string]string{
		labels.RegistryaasInstanceLabel: labelTrueValue,
		labels.RegistryaasUsernameLabel: username,
		labels.RegistryaasPortLabel:     strconv.Itoa(int(port)),
	}
}

// registryLoginUrl addresses the registry API the way the runtime says Velez
// reaches it: over a shared Docker network or through the published port.
func registryLoginUrl(address string) string {
	loginUrl := url.URL{
		Scheme: "http",
		Host:   address,
		Path:   registryApiPath,
	}

	return loginUrl.String()
}

func pingRegistryLogin(ctx context.Context, loginUrl string, login registryLogin) error {
	pingCtx, cancel := context.WithTimeout(ctx, registeredRegistryTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(pingCtx, http.MethodGet, loginUrl, http.NoBody)
	if err != nil {
		return rerrors.Wrap(err, "error building registry login request")
	}

	request.SetBasicAuth(login.username, login.password)

	httpClient := &http.Client{Timeout: registeredRegistryTimeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return rerrors.Wrap(rerrors.Join(user_errors.ErrRegistryLoginFailed, err))
	}

	defer func() {
		closeErr := response.Body.Close()
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("error closing registry response body when verifying login")
		}
	}()

	if response.StatusCode != http.StatusOK {
		return rerrors.Wrap(user_errors.ErrRegistryLoginFailed, fmt.Sprintf("unexpected status: %d", response.StatusCode))
	}

	return nil
}

func loadRegistryLogin(
	ctx context.Context, store secrets.Store, registry registerRegistryAccessor,
) (registryLogin, error) {
	login := registryLogin{username: registry.GetRegistryUsername()}

	owner := registry.GetRegistryPendingSecretOwner()
	if owner != "" {
		password, err := readOptionalSecret(ctx, store, RegisteredRegistryPendingSecretRef(owner))
		if err != nil {
			return registryLogin{}, rerrors.Wrap(err)
		}

		login.password = password
	}

	if login.username == "" || login.password == "" {
		return registryLogin{}, rerrors.Wrap(user_errors.ErrRegistryCredentialsRequired)
	}

	return login, nil
}

// verifyRegistryLoginJob logs into the registry with the request's login
// before anything is written or restarted. The password is never put into the
// payload. A container without htpasswd auth needs no login: the job then
// clears the username so no later job stamps it.
type verifyRegistryLoginJob struct {
	runtimes container_runtime.RuntimeResolver
	secrets  secrets.Store

	req      registerRequestAccessor
	registry registerRegistryAccessor
	ctx      registerRegistrySetter
}

func (j *verifyRegistryLoginJob) Do(ctx context.Context) error {
	info, err := inspectRegisteredContainer(ctx, j.runtimes, j.req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	containerEnv := parser.ToDockerEnv(info.Config.Env)

	if !isRegistryAuthConfigured(containerEnv) {
		j.ctx.SetRegistryUsername("")

		return nil
	}

	login, err := loadRegistryLogin(ctx, j.secrets, j.registry)
	if err != nil {
		return rerrors.Wrap(err)
	}

	containerPort := int(registryContainerPort(containerEnv))

	address, err := registeredContainerAddress(ctx, j.runtimes, j.req, info, containerPort)
	if err != nil {
		return rerrors.Wrap(err)
	}

	loginUrl := registryLoginUrl(address)

	err = pingRegistryLogin(ctx, loginUrl, login)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

// storeRegistrySecretJob writes the final registryaas secret from the
// verified login and drops the pending one.
type storeRegistrySecretJob struct {
	secrets secrets.Store

	req      registerRequestAccessor
	registry registerRegistryAccessor
}

func (j *storeRegistrySecretJob) Do(ctx context.Context) error {
	if !hasRegistryLogin(j.registry) {
		return nil
	}

	login, err := loadRegistryLogin(ctx, j.secrets, j.registry)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.secrets.Put(ctx, registryInstanceSecretRef(j.req.GetServiceName()), login.password)
	if err != nil {
		return rerrors.Wrap(err, "error storing registry password")
	}

	err = deletePendingRegistrySecret(ctx, j.secrets, j.registry.GetRegistryPendingSecretOwner())
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

// upsertRegistryRowJob runs right after the bind transaction: RegistryInstances
// has no WithTx, and the upsert is idempotent.
type upsertRegistryRowJob struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver

	req      registerRequestAccessor
	registry registerRegistryAccessor
	service  registeredServiceAccessor
}

func (j *upsertRegistryRowJob) Do(ctx context.Context) error {
	info, err := inspectRegisteredContainer(ctx, j.runtimes, j.req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	upsertReq := domain.UpsertRegistryInstanceReq{
		ServiceId: j.service.GetServiceId(),
		Port:      registryInspectedPort(info),
		Username:  j.registry.GetRegistryUsername(),
		SecretRef: registryInstanceSecretRef(j.req.GetServiceName()).String(),
	}

	_, err = j.dataStorage.RegistryInstances().UpsertRegistryInstance(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting registry instance row")
	}

	return nil
}

// applyRegistryOverlay must run after applyRegisterOverrides: the stamped
// port is the one the recreated container publishes.
func (j *recreateWithLabelsJob) applyRegistryOverlay(
	payload *velez_api.UpgradeSmerdTaskPayload, info container.InspectResponse,
) {
	containerPort := registryContainerPort(parser.ToDockerEnv(info.Config.Env))
	port := registryPortFromMappings(payload.GetPortsOverride().GetPorts(), containerPort)

	maps.Copy(payload.GetExtraLabels(), registeredRegistryLabels(j.registry.GetRegistryUsername(), port))
}
