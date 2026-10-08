package jobs

import (
	"encoding/json"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	pgDataSource = "/srv/pgdata"
	pgDataTarget = "/var/lib/postgresql/data"
)

func newPgMounts() []container.MountPoint {
	return []container.MountPoint{
		{Type: mount.TypeBind, Source: pgDataSource, Destination: pgDataTarget},
		{Type: mount.TypeVolume, Name: "cache", Destination: "/cache"},
		{Type: mount.TypeTmpfs, Destination: "/run"},
	}
}

func Test_DeriveLinkedVolumeName_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		service string
		target  string
		want    string
	}{
		{"nested path", "pg", pgDataTarget, "pg_var_lib_postgresql_data"},
		{"invalid chars collapse", "my svc", "/data dir", "my_svc_data_dir"},
		{"root target", "pg", "/", "pg_"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, deriveLinkedVolumeName(tc.service, tc.target))
		})
	}
}

func Test_ResolveBindMountLinks_DerivesNameAndSkipsOtherMountKinds(t *testing.T) {
	t.Parallel()

	links := []*bindMountLink{{Source: pgDataSource}}

	linked, err := resolveBindMountLinks("pg", newPgMounts(), links)

	require.NoError(t, err)
	require.Equal(t, []linkedBindMount{{
		source:     pgDataSource,
		target:     pgDataTarget,
		volumeName: "pg_var_lib_postgresql_data",
	}}, linked)
}

func Test_ResolveBindMountLinks_ExplicitVolumeNameWins(t *testing.T) {
	t.Parallel()

	links := []*bindMountLink{{Source: pgDataSource, VolumeName: toolbox.ToPtr("pgdata")}}

	linked, err := resolveBindMountLinks("pg", newPgMounts(), links)

	require.NoError(t, err)
	require.Equal(t, "pgdata", linked[0].volumeName)
}

func Test_ResolveBindMountLinks_UnlinkedBindMountIsRefusedNamingIt(t *testing.T) {
	t.Parallel()

	_, err := resolveBindMountLinks("pg", newPgMounts(), nil)

	require.ErrorIs(t, err, user_errors.ErrBindMountNotLinked)
	require.Contains(t, err.Error(), "/srv/pgdata -> /var/lib/postgresql/data")
}

func Test_ResolveBindMountLinks_LinkMatchingNoBindMountIsRefused(t *testing.T) {
	t.Parallel()

	links := []*bindMountLink{{Source: pgDataSource}, {Source: "/srv/other"}}

	_, err := resolveBindMountLinks("pg", newPgMounts(), links)

	require.ErrorIs(t, err, user_errors.ErrBindMountLinkUnknown)
	require.Contains(t, err.Error(), "/srv/other")
}

func Test_ResolveBindMountLinks_NoBindMountsNoLinksIsFine(t *testing.T) {
	t.Parallel()

	linked, err := resolveBindMountLinks("pg", nil, nil)

	require.NoError(t, err)
	require.Empty(t, linked)
}

func Test_BindVolumeRequest_BacksVolumeByHostDirectory(t *testing.T) {
	t.Parallel()

	got := bindVolumeRequest(linkedBindMount{source: pgDataSource, volumeName: "pgdata"})

	require.Equal(t, "pgdata", got.Name)
	require.Equal(t, "local", got.Driver)
	require.Equal(t, map[string]string{"type": "none", "o": "bind", "device": pgDataSource}, got.DriverOpts)
}

type fakeRegisterUpgradeRequest struct {
	serviceName string
	keepPorts   bool
	ports       []*velez_api.Port
	links       []*bindMountLink
}

func (f fakeRegisterUpgradeRequest) GetServiceName() string { return f.serviceName }

func (f fakeRegisterUpgradeRequest) GetKeepPortMapping() bool { return f.keepPorts }

func (f fakeRegisterUpgradeRequest) GetPorts() []*velez_api.Port { return f.ports }

func (f fakeRegisterUpgradeRequest) GetBindMountLinks() []*bindMountLink { return f.links }

func newPublishedContainerInfo() container.InspectResponse {
	pgPort, err := nat.NewPort("tcp", "5432")
	if err != nil {
		panic(err)
	}

	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				PortBindings: nat.PortMap{pgPort: {{HostPort: "15432"}}},
			},
		},
		Mounts: newPgMounts(),
	}
}

func Test_ApplyRegisterOverrides_KeepPortMappingStopsOldFirstAndReusesHostPorts(t *testing.T) {
	t.Parallel()

	payload := &velez_api.UpgradeSmerdTaskPayload{}
	req := fakeRegisterUpgradeRequest{serviceName: "pg", keepPorts: true}

	err := applyRegisterOverrides(payload, newPublishedContainerInfo(), req)

	require.NoError(t, err)
	require.True(t, payload.GetStopOldFirst())

	ports := payload.GetPortsOverride().GetPorts()
	require.Len(t, ports, 1)
	require.Equal(t, uint32(5432), ports[0].GetServicePortNumber())
	require.Equal(t, uint32(15432), ports[0].GetExposedTo())
	require.Nil(t, payload.GetVolumesOverride())
}

func Test_ApplyRegisterOverrides_RequestPortsReplaceInspectedOnes(t *testing.T) {
	t.Parallel()

	payload := &velez_api.UpgradeSmerdTaskPayload{}
	requested := []*velez_api.Port{{ServicePortNumber: 5432}}
	req := fakeRegisterUpgradeRequest{serviceName: "pg", ports: requested}

	err := applyRegisterOverrides(payload, newPublishedContainerInfo(), req)

	require.NoError(t, err)
	require.True(t, payload.GetStopOldFirst(), "the inspected container has mounts the replacement shares")
	require.Equal(t, requested, payload.GetPortsOverride().GetPorts())
}

func Test_ApplyRegisterOverrides_NoPortsMeansPresentEmptyOverride(t *testing.T) {
	t.Parallel()

	payload := &velez_api.UpgradeSmerdTaskPayload{}
	req := fakeRegisterUpgradeRequest{serviceName: "pg"}

	err := applyRegisterOverrides(payload, newPublishedContainerInfo(), req)

	require.NoError(t, err)
	require.NotNil(t, payload.GetPortsOverride())
	require.Empty(t, payload.GetPortsOverride().GetPorts())
}

func Test_ApplyRegisterOverrides_LinksReplaceBindMountsAndKeepNamedVolumes(t *testing.T) {
	t.Parallel()

	payload := &velez_api.UpgradeSmerdTaskPayload{}
	req := fakeRegisterUpgradeRequest{serviceName: "pg", links: []*bindMountLink{{Source: pgDataSource}}}

	err := applyRegisterOverrides(payload, newPublishedContainerInfo(), req)

	require.NoError(t, err)

	volumes := payload.GetVolumesOverride().GetVolumes()
	require.Len(t, volumes, 2)
	require.Equal(t, "cache", volumes[0].GetVolumeName())
	require.Equal(t, "pg_var_lib_postgresql_data", volumes[1].GetVolumeName())
	require.Equal(t, pgDataTarget, volumes[1].GetContainerPath())
}

func Test_ApplyUpgradeOverrides_UnsetKeepsCapturedSettings(t *testing.T) {
	t.Parallel()

	captured := []*velez_api.Port{{ServicePortNumber: 80}}
	req := &velez_api.CreateSmerd_Request{Settings: &velez_api.Container_Settings{Ports: captured}}

	applyUpgradeOverrides(req, &velez_api.UpgradeSmerdTaskPayload{})

	require.Equal(t, captured, req.GetSettings().GetPorts())
}

func Test_ApplyUpgradeOverrides_PresentEmptyClearsAndValuesReplace(t *testing.T) {
	t.Parallel()

	req := &velez_api.CreateSmerd_Request{Settings: &velez_api.Container_Settings{
		Ports:   []*velez_api.Port{{ServicePortNumber: 80}},
		Volumes: []*velez_api.Volume{{VolumeName: testStaleOld}},
	}}
	replacement := []*velez_api.Volume{{VolumeName: "new", ContainerPath: "/data"}}
	overrides := &velez_api.UpgradeSmerdTaskPayload{
		PortsOverride:   &velez_api.UpgradeSmerdTaskPayload_PortsOverride{},
		VolumesOverride: &velez_api.UpgradeSmerdTaskPayload_VolumesOverride{Volumes: replacement},
	}

	applyUpgradeOverrides(req, overrides)

	require.Empty(t, req.GetSettings().GetPorts())
	require.Equal(t, replacement, req.GetSettings().GetVolumes())
}

func Test_UpgradeSmerdTaskPayload_EmptyPortsOverrideSurvivesJsonRoundTrip(t *testing.T) {
	t.Parallel()

	payload := &velez_api.UpgradeSmerdTaskPayload{
		StopOldFirst:  true,
		PortsOverride: &velez_api.UpgradeSmerdTaskPayload_PortsOverride{},
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	restored := &velez_api.UpgradeSmerdTaskPayload{}

	err = json.Unmarshal(raw, restored)
	require.NoError(t, err)

	require.True(t, restored.GetStopOldFirst())
	require.NotNil(t, restored.GetPortsOverride())
	require.Nil(t, restored.GetVolumesOverride())
}
