package verv_services

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testContID1      = "cont-1"
	testContID2      = "cont-2"
	testEnvSuffixStg = "stg"
)

// fakeCascadeDocker records what the cascade asked Docker to list and remove.
type fakeCascadeDocker struct {
	containers []container.Summary
	listSuffix string
	listErr    error

	removed   []string
	removeErr error
}

func (f *fakeCascadeDocker) ListContainers(
	_ context.Context, _ *velez_api.ListSmerds_Request, suffix string,
) ([]container.Summary, error) {
	f.listSuffix = suffix

	return f.containers, f.listErr
}

func (f *fakeCascadeDocker) Remove(_ context.Context, uuid string) error {
	f.removed = append(f.removed, uuid)

	return f.removeErr
}

type fakeCascadeNetVol struct {
	networks    []network.Summary
	networkList network.ListOptions
	networkErr  error

	volumes   []*volume.Volume
	volumeErr error

	removedNetworks []string
	removedVolumes  []string
}

func (f *fakeCascadeNetVol) NetworkList(
	_ context.Context, options network.ListOptions,
) ([]network.Summary, error) {
	f.networkList = options

	return f.networks, f.networkErr
}

func (f *fakeCascadeNetVol) NetworkRemove(_ context.Context, networkID string) error {
	f.removedNetworks = append(f.removedNetworks, networkID)

	return nil
}

func (f *fakeCascadeNetVol) VolumeList(
	_ context.Context, _ volume.ListOptions,
) (volume.ListResponse, error) {
	return volume.ListResponse{Volumes: f.volumes}, f.volumeErr
}

func (f *fakeCascadeNetVol) VolumeRemove(_ context.Context, volumeID string, _ bool) error {
	f.removedVolumes = append(f.removedVolumes, volumeID)

	return nil
}

func TestCascadeRemoveEnvironmentResources_RemovesContainersNetworksVolumes(t *testing.T) {
	docker := &fakeCascadeDocker{
		containers: []container.Summary{{ID: testContID1}, {ID: testContID2}},
	}
	netVol := &fakeCascadeNetVol{
		networks: []network.Summary{{ID: "net-1"}},
		volumes:  []*volume.Volume{{Name: "vol-1"}},
	}

	err := cascadeRemoveEnvironmentResources(context.Background(), docker, netVol, testEnvSuffixStg)
	require.NoError(t, err)

	require.Equal(t, testEnvSuffixStg, docker.listSuffix)
	require.Equal(t, []string{testContID1, testContID2}, docker.removed)
	require.Equal(t, []string{"net-1"}, netVol.removedNetworks)
	require.Equal(t, []string{"vol-1"}, netVol.removedVolumes)

	// Networks/volumes must be selected by the suffix label, never by name.
	require.True(t, netVol.networkList.Filters.ExactMatch("label", labels.SuffixLabel+"=stg"))
}

// An empty suffix would match every Velez-created container on the node (they
// all carry labels.SuffixLabel, set to "" when unconfigured), so the cascade
// must refuse it outright rather than wiping the host.
func TestCascadeRemoveEnvironmentResources_EmptySuffixRefused(t *testing.T) {
	docker := &fakeCascadeDocker{containers: []container.Summary{{ID: testContID1}}}
	netVol := &fakeCascadeNetVol{}

	err := cascadeRemoveEnvironmentResources(context.Background(), docker, netVol, "")
	require.Error(t, err)
	require.Empty(t, docker.removed)
	require.Empty(t, netVol.removedNetworks)
	require.Empty(t, netVol.removedVolumes)
}

// A container removal failure must not stop the sweep - every remaining
// container is still attempted - but it must still surface as an error so the
// DB row isn't dropped on top of a half-cleaned environment.
func TestCascadeRemoveEnvironmentResources_ContainerFailureStillSweepsAndErrors(t *testing.T) {
	docker := &fakeCascadeDocker{
		containers: []container.Summary{{ID: testContID1}, {ID: testContID2}},
		removeErr:  rerrors.New("boom"),
	}
	netVol := &fakeCascadeNetVol{}

	err := cascadeRemoveEnvironmentResources(context.Background(), docker, netVol, testEnvSuffixStg)
	require.Error(t, err)
	require.Equal(t, []string{testContID1, testContID2}, docker.removed)
	require.Empty(t, netVol.removedNetworks, "networks are not touched once containers failed")
}

func TestCascadeRemoveEnvironmentResources_ListFailurePropagates(t *testing.T) {
	docker := &fakeCascadeDocker{listErr: rerrors.New("docker down")}

	err := cascadeRemoveEnvironmentResources(context.Background(), docker, &fakeCascadeNetVol{}, testEnvSuffixStg)
	require.Error(t, err)
}
