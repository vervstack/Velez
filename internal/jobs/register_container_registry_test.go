package jobs

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testRegistryUser     = "verv"
	testRegistryPassword = "s3cret"
)

func Test_IsRegistryAuthConfigured_DetectsHtpasswdEnv(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"no env", map[string]string{}, false},
		{"auth driver htpasswd", map[string]string{envRegistryAuth: "htpasswd"}, true},
		{"htpasswd path only", map[string]string{envRegistryAuthHtpasswdPath: "/auth/htpasswd"}, true},
		{"other auth driver", map[string]string{envRegistryAuth: "token"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, isRegistryAuthConfigured(tc.env))
		})
	}
}

func Test_RegistryContainerPort_ReadsHttpAddrOrFallsBack(t *testing.T) {
	t.Parallel()

	require.EqualValues(t, registryaasContainerPort, registryContainerPort(map[string]string{}))

	badAddr := map[string]string{registryHttpAddrEnvVar: "bad"}

	require.EqualValues(t, registryaasContainerPort, registryContainerPort(badAddr))
	require.EqualValues(t, 5443, registryContainerPort(map[string]string{registryHttpAddrEnvVar: "0.0.0.0:5443"}))
	require.EqualValues(t, 5444, registryContainerPort(map[string]string{registryHttpAddrEnvVar: ":5444"}))
}

func Test_RegistryPortFromMappings_PrefersPublishedHostPort(t *testing.T) {
	t.Parallel()

	published := &velez_api.Port{ServicePortNumber: registryaasContainerPort, ExposedTo: toolbox.ToPtr(uint32(30019))}
	unpublished := &velez_api.Port{ServicePortNumber: registryaasContainerPort}

	require.EqualValues(t, 30019, registryPortFromMappings([]*velez_api.Port{published}, registryaasContainerPort))

	unpublishedPorts := []*velez_api.Port{unpublished}

	require.EqualValues(t, registryaasContainerPort, registryPortFromMappings(unpublishedPorts, registryaasContainerPort))
	require.EqualValues(t, registryaasContainerPort, registryPortFromMappings(nil, registryaasContainerPort))
}

func Test_RegisteredRegistryLabels_RoundTripThroughContainerDerived(t *testing.T) {
	t.Parallel()

	stamped := registeredRegistryLabels(testRegistryUser, 30019)

	instance := container_derived.RegistryInstance("reg", time.Time{}, stamped)

	require.Equal(t, testRegistryUser, instance.Username)
	require.EqualValues(t, 30019, instance.Port)
	require.EqualValues(t, 0, instance.UiPort)
}

func Test_RegistryLoginUrl_UsesRuntimeAddress(t *testing.T) {
	t.Parallel()

	require.Equal(t, "http://172.17.0.7:5000/v2/", registryLoginUrl("172.17.0.7:5000"))
	require.Equal(t, "http://sc:30019/v2/", registryLoginUrl("sc:30019"))
}

func Test_PingRegistryLogin_AcceptsOnlyAuthorizedLogin(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != testRegistryUser || password != testRegistryPassword {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	right := registryLogin{username: testRegistryUser, password: testRegistryPassword}
	wrong := registryLogin{username: testRegistryUser, password: "wrong"}

	require.NoError(t, pingRegistryLogin(t.Context(), server.URL, right))
	require.ErrorIs(t, pingRegistryLogin(t.Context(), server.URL, wrong), user_errors.ErrRegistryLoginFailed)
}
