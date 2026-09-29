// Package container_derived rebuilds the AAS satellite rows (pg instance,
// runner, registry instance) from a labelled container. Single-node/dev
// local_storage treats the container as the system of record and the
// enable_statefull backfill copies those rows into Postgres, so both share
// this one conversion.
package container_derived

import (
	"strconv"
	"strings"
	"time"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	// The scope+key halves of the secret refs the pgaas and registryaas
	// services store their generated passwords under. Owner is the instance name.
	PgaasSecretScope       = "pgaas"
	PgaasSecretKey         = "password"
	RegistryaasSecretScope = "registryaas"
	RegistryaasSecretKey   = "password"

	// The container env vars pgaas.buildDeployRequest writes an instance's
	// facts into.
	PgaasEnvDbName   = "POSTGRES_DB"
	PgaasEnvUsername = "POSTGRES_USER"
	PgaasEnvPassword = "POSTGRES_PASSWORD"

	PgaasDefaultPort = 5432

	// RegistryaasDefaultPort is the container-internal port, used only when a
	// container predates labels.RegistryaasPortLabel.
	RegistryaasDefaultPort = 5000

	// RunnerRegistrationTokenEnvVar mirrors create_runner.go's const of the
	// same name - the env var carrying the minted registration token into the
	// runner container.
	RunnerRegistrationTokenEnvVar = "VELEZ_RUNNER_REGISTRATION_TOKEN"
)

func PgInstanceSecretRef(name string) domain.SecretRef {
	return domain.SecretRef{Scope: PgaasSecretScope, Owner: name, Key: PgaasSecretKey}
}

func RegistryInstanceSecretRef(name string) domain.SecretRef {
	return domain.SecretRef{Scope: RegistryaasSecretScope, Owner: name, Key: RegistryaasSecretKey}
}

// ServiceName is the service a container belongs to: its VervServiceLabel
// value, falling back to its first container name.
func ServiceName(containerLabels map[string]string, containerNames []string) string {
	name := containerLabels[labels.VervServiceLabel]
	if name == "" && len(containerNames) != 0 {
		name = strings.TrimPrefix(containerNames[0], "/")
	}

	return name
}

// EnvValue returns the value of key in a Docker "KEY=VALUE" env slice, or "".
func EnvValue(env []string, key string) string {
	prefix := key + "="

	for _, entry := range env {
		value, ok := strings.CutPrefix(entry, prefix)
		if ok {
			return value
		}
	}

	return ""
}

// PgInstance derives the row from the POSTGRES_* env pgaas wrote at deploy
// time. ServiceId is left for the caller: it is storage-specific.
func PgInstance(name string, created time.Time, env []string) domain.PgInstance {
	return domain.PgInstance{
		DbName:    EnvValue(env, PgaasEnvDbName),
		Username:  EnvValue(env, PgaasEnvUsername),
		SecretRef: PgInstanceSecretRef(name).String(),
		Port:      PgaasDefaultPort,
		CreatedAt: created,
		UpdatedAt: created,
	}
}

// Runner derives the row from the runner container's labels. Concurrent and
// ServiceID are left for the caller.
func Runner(name string, created time.Time, containerLabels map[string]string) domain.Runner {
	return domain.Runner{
		Provider:  containerLabels[labels.RunnerProviderLabel],
		Scope:     containerLabels[labels.RunnerScopeLabel],
		Target:    containerLabels[labels.RunnerTargetLabel],
		Labels:    splitRunnerLabels(containerLabels[labels.RunnerLabelsLabel]),
		SecretRef: domain.RunnerAccessTokenSecretRef(name).String(),
		BaseUrl:   containerLabels[labels.RunnerBaseUrlLabel],
		CreatedAt: created,
		UpdatedAt: created,
	}
}

func IsGitlabRunner(runner domain.Runner) bool {
	return runner.Provider == pb.RunnerProvider_GITLAB.String()
}

// RegistryInstance derives the row from the registry container's labels.
// ServiceId is left for the caller.
func RegistryInstance(name string, created time.Time, containerLabels map[string]string) domain.RegistryInstance {
	return domain.RegistryInstance{
		Port:      portFromLabel(containerLabels[labels.RegistryaasPortLabel]),
		UiPort:    uiPortFromLabel(containerLabels[labels.RegistryaasUiPortLabel]),
		Username:  containerLabels[labels.RegistryaasUsernameLabel],
		SecretRef: RegistryInstanceSecretRef(name).String(),
		CreatedAt: created,
		UpdatedAt: created,
	}
}

// splitRunnerLabels is the inverse of strings.Join(labels, ",") - an empty
// value means no labels were set, not one empty-string label.
func splitRunnerLabels(value string) []string {
	if value == "" {
		return nil
	}

	return strings.Split(value, ",")
}

// uiPortFromLabel falls back to the 0 "not provisioned yet" sentinel on an
// empty or malformed label.
func uiPortFromLabel(value string) int32 {
	port, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}

	return int32(port) //nolint:gosec
}

// portFromLabel falls back to RegistryaasDefaultPort only for a container
// created before labels.RegistryaasPortLabel existed.
func portFromLabel(value string) int32 {
	port, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return RegistryaasDefaultPort
	}

	return int32(port) //nolint:gosec
}
