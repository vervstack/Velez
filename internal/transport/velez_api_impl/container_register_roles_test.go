package velez_api_impl

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	testRolesRootId     = "aaaa1111"
	testRolesSidecarId  = "bbbb2222"
	testRolesForeignId  = "cccc3333"
	testRolesForeignOwn = "dddd4444"
	testRolesThirdId    = "eeee5555"
)

func newRolesContainers() []*velez_api.DockerContainer {
	foreignOwner := testRolesForeignOwn

	return []*velez_api.DockerContainer{
		{Id: testRolesRootId + "ff"},
		{Id: testRolesSidecarId + "ff", NetworkOwnerContainerId: ptrOf(testRolesRootId + "ff")},
		{Id: testRolesThirdId + "ff", NetworkOwnerContainerId: ptrOf(testRolesRootId + "ff")},
		{Id: testRolesForeignId + "ff", NetworkOwnerContainerId: &foreignOwner},
	}
}

func ptrOf(v string) *string {
	return &v
}

func Test_SelectRegisterRoles_ServiceIsRootKeepsRootAsService(t *testing.T) {
	t.Parallel()

	refs := registerRoleRefs{Requested: testRolesRootId, Service: testRolesRootId}

	roles, err := selectRegisterRoles(newRolesContainers(), refs)
	require.NoError(t, err)
	require.Equal(t, registerRoles{ServiceId: testRolesRootId + "ff"}, roles)
}

func Test_SelectRegisterRoles_SidecarAsServiceResolvesFullIds(t *testing.T) {
	t.Parallel()

	refs := registerRoleRefs{Requested: testRolesRootId, Service: testRolesSidecarId}

	roles, err := selectRegisterRoles(newRolesContainers(), refs)
	require.NoError(t, err)
	require.Equal(t, registerRoles{RootId: testRolesRootId + "ff", ServiceId: testRolesSidecarId + "ff"}, roles)
}

func Test_SelectRegisterRoles_RejectsServiceOutsideGroup(t *testing.T) {
	t.Parallel()

	for _, serviceRef := range []string{testRolesForeignId, "unknown"} {
		refs := registerRoleRefs{Requested: testRolesRootId, Service: serviceRef}

		_, err := selectRegisterRoles(newRolesContainers(), refs)
		require.ErrorIs(t, err, user_errors.ErrRegisterServiceNotInGroup)
	}
}

func Test_SelectRegisterRoles_NewRootIsServiceByDefault(t *testing.T) {
	t.Parallel()

	refs := registerRoleRefs{Requested: testRolesRootId, Root: testRolesSidecarId}

	roles, err := selectRegisterRoles(newRolesContainers(), refs)
	require.NoError(t, err)

	want := registerRoles{ServiceId: testRolesSidecarId + "ff", PreviousRootId: testRolesRootId + "ff"}
	require.Equal(t, want, roles)
	require.True(t, roles.isRootSwitched())
}

func Test_SelectRegisterRoles_NewRootWithOtherService(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		serviceRef string
		wantSvc    string
	}{
		{"old root as service", testRolesRootId, testRolesRootId + "ff"},
		{"third member as service", testRolesThirdId, testRolesThirdId + "ff"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			refs := registerRoleRefs{Requested: testRolesRootId, Root: testRolesSidecarId, Service: tc.serviceRef}

			roles, err := selectRegisterRoles(newRolesContainers(), refs)
			require.NoError(t, err)

			want := registerRoles{
				RootId:         testRolesSidecarId + "ff",
				ServiceId:      tc.wantSvc,
				PreviousRootId: testRolesRootId + "ff",
			}
			require.Equal(t, want, roles)
		})
	}
}

func Test_SelectRegisterRoles_RootNamingTheRequestedContainerIsNoSwitch(t *testing.T) {
	t.Parallel()

	refs := registerRoleRefs{Requested: testRolesRootId, Root: testRolesRootId}

	roles, err := selectRegisterRoles(newRolesContainers(), refs)
	require.NoError(t, err)
	require.False(t, roles.isRootSwitched())
	require.Equal(t, registerRoles{ServiceId: testRolesRootId + "ff"}, roles)
}

func Test_SelectRegisterRoles_RejectsRootOutsideGroup(t *testing.T) {
	t.Parallel()

	for _, rootRef := range []string{testRolesForeignId, "unknown"} {
		refs := registerRoleRefs{Requested: testRolesRootId, Root: rootRef}

		_, err := selectRegisterRoles(newRolesContainers(), refs)
		require.ErrorIs(t, err, user_errors.ErrRegisterRootNotInGroup)
	}
}

func Test_ValidateRecreateFreeRequest(t *testing.T) {
	t.Parallel()

	tag := "1.0"

	cases := []struct {
		name    string
		req     *velez_api.RegisterContainer_Request
		isValid bool
	}{
		{"generic request", &velez_api.RegisterContainer_Request{}, true},
		{"explicit generic pattern", &velez_api.RegisterContainer_Request{
			Pattern: &velez_api.RegisterContainer_Request_Generic{
				Generic: &velez_api.RegisterContainer_Request_GenericPattern{},
			},
		}, true},
		{"pg pattern", &velez_api.RegisterContainer_Request{
			Pattern: &velez_api.RegisterContainer_Request_Pg{Pg: &velez_api.RegisterContainer_Request_PgPattern{}},
		}, false},
		{"ports", &velez_api.RegisterContainer_Request{Ports: []*velez_api.Port{{}}}, false},
		{"image tag", &velez_api.RegisterContainer_Request{ImageTag: &tag}, false},
		{"keep port mapping", &velez_api.RegisterContainer_Request{KeepPortMapping: true}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateRecreateFreeRequest(tc.req, user_errors.ErrRegisterServiceContainerUnsupported)
			if tc.isValid {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, user_errors.ErrRegisterServiceContainerUnsupported)
		})
	}
}
