package local_storage

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/tests/test_helper"
)

const (
	dockerSocketBind = dockerSocketPath + ":" + dockerSocketPath
)

func createSecretsTestContainer(t *testing.T, name string, hostCfg *container.HostConfig) {
	t.Helper()

	cli := test_helper.NewRealDockerAPI(t)
	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	cfg := &container.Config{
		Image:  test_helper.HelloWorldAppImage,
		Labels: map[string]string{labels.VervServiceLabel: name},
	}

	created, err := cli.ContainerCreate(context.Background(), cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	t.Cleanup(func() {
		test_helper.RemoveContainer(t, cli, created.ID)
	})
}

func Test_DockerSecrets_SocketGrantDerivedFromSocketBind(t *testing.T) {
	t.Parallel()

	name := test_helper.UniqueName(t, "socket-bound")
	hostCfg := &container.HostConfig{Binds: []string{dockerSocketBind}}
	createSecretsTestContainer(t, name, hostCfg)

	s := newSecretsStorage(test_helper.NewRealDocker(t))
	ref := domain.DockerSocketGrantSecretRef(name)

	value, err := s.GetSecret(context.Background(), ref)
	require.NoError(t, err)
	require.Equal(t, "true", value)
}

func Test_DockerSecrets_SocketGrantNotDerivedWithoutSocketBind(t *testing.T) {
	t.Parallel()

	name := test_helper.UniqueName(t, "socket-unbound")
	createSecretsTestContainer(t, name, nil)

	s := newSecretsStorage(test_helper.NewRealDocker(t))
	ref := domain.DockerSocketGrantSecretRef(name)

	_, err := s.GetSecret(context.Background(), ref)
	require.ErrorIs(t, err, user_errors.ErrStorageNotFound)
}

func Test_DockerSecrets_SocketGrantFallsBackToMemoryBeforeContainerExists(t *testing.T) {
	t.Parallel()

	name := test_helper.UniqueName(t, "socket-pending")
	s := newSecretsStorage(test_helper.NewRealDocker(t))
	ref := domain.DockerSocketGrantSecretRef(name)

	err := s.PutSecret(context.Background(), ref, "true")
	require.NoError(t, err)

	value, err := s.GetSecret(context.Background(), ref)
	require.NoError(t, err)
	require.Equal(t, "true", value)
}

func Test_DockerSecrets_IsolationDerivedFromPrivilegedContainer(t *testing.T) {
	t.Parallel()

	name := test_helper.UniqueName(t, "isolation-privileged")
	hostCfg := &container.HostConfig{Privileged: true}
	createSecretsTestContainer(t, name, hostCfg)

	s := newSecretsStorage(test_helper.NewRealDocker(t))
	ref := domain.ContainerIsolationSecretRef(name)

	value, err := s.GetSecret(context.Background(), ref)
	require.NoError(t, err)
	require.Equal(t, domain.ContainerIsolationPrivileged, value)
}

func Test_DockerSecrets_IsolationNotDerivedFromPlainContainer(t *testing.T) {
	t.Parallel()

	name := test_helper.UniqueName(t, "isolation-plain")
	createSecretsTestContainer(t, name, nil)

	s := newSecretsStorage(test_helper.NewRealDocker(t))
	ref := domain.ContainerIsolationSecretRef(name)

	_, err := s.GetSecret(context.Background(), ref)
	require.ErrorIs(t, err, user_errors.ErrStorageNotFound)
}
