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
)

func newRolesContainers() []*velez_api.DockerContainer {
	foreignOwner := testRolesForeignOwn

	return []*velez_api.DockerContainer{
		{Id: testRolesRootId + "ff"},
		{Id: testRolesSidecarId + "ff", NetworkOwnerContainerId: ptrOf(testRolesRootId + "ff")},
		{Id: testRolesForeignId + "ff", NetworkOwnerContainerId: &foreignOwner},
	}
}

func ptrOf(v string) *string {
	return &v
}

func Test_SelectRegisterRoles_ServiceIsRootKeepsRootAsService(t *testing.T) {
	t.Parallel()

	roles, err := selectRegisterRoles(newRolesContainers(), testRolesRootId, testRolesRootId)
	require.NoError(t, err)
	require.Equal(t, registerRoles{ServiceId: testRolesRootId}, roles)
}

func Test_SelectRegisterRoles_SidecarAsServiceResolvesFullIds(t *testing.T) {
	t.Parallel()

	roles, err := selectRegisterRoles(newRolesContainers(), testRolesRootId, testRolesSidecarId)
	require.NoError(t, err)
	require.Equal(t, registerRoles{RootId: testRolesRootId + "ff", ServiceId: testRolesSidecarId + "ff"}, roles)
}

func Test_SelectRegisterRoles_RejectsServiceOutsideGroup(t *testing.T) {
	t.Parallel()

	for _, serviceRef := range []string{testRolesForeignId, "unknown"} {
		_, err := selectRegisterRoles(newRolesContainers(), testRolesRootId, serviceRef)
		require.ErrorIs(t, err, user_errors.ErrRegisterServiceNotInGroup)
	}
}

func Test_ValidateServiceContainerRequest(t *testing.T) {
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

			err := validateServiceContainerRequest(tc.req)
			if tc.isValid {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, user_errors.ErrRegisterServiceContainerUnsupported)
		})
	}
}
