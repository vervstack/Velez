package vervonomicon

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

const (
	testSvcName         = "svc"
	testDescriptorImage = "registry/example:from-descriptor"
)

func fullDescriptor() verv.Descriptor {
	return verv.Descriptor{
		Index: verv.Index{
			Version: "1",
			Service: verv.Service{
				Name: "my-service",
				Repo: "https://github.com/example/my-service",
				Tags: []string{"backend", "critical"},
			},
			Box: boxSmall,
		},
		Deployment: verv.Deployment{
			App: verv.App{
				Command:       "serve --port 8080",
				UseImagePorts: false,
				Ports: []verv.Port{
					{Port: 8080, Protocol: verv.ProtocolTcp, ExposeTo: 80},
					{Port: 9090, Protocol: verv.ProtocolUdp},
				},
				Volumes: []verv.VolumeMount{
					{Name: "data", Path: "/var/lib/data"},
				},
				Env:    map[string]string{"LOG_LEVEL": "info"},
				Labels: map[string]string{"team": "platform"},
				Healthcheck: verv.Healthcheck{
					Command:        "curl -f localhost:8080/health",
					IntervalSecond: 10,
					TimeoutSecond:  5,
					Retries:        3,
				},
				Restart:             verv.RestartPolicyOnFailure,
				RestartFailureCount: 5,
				AutoUpgrade:         true,
			},
		},
	}
}

func TestBoxResolver_ResolveRequest_FullFieldMapping(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	descriptor := fullDescriptor()

	got, err := resolver.ResolveRequest(context.Background(), descriptor, "production", "")
	require.NoError(t, err)

	require.Equal(t, "my-service", got.GetName())
	require.Equal(t, "https://github.com/example/my-service", got.GetRepo())
	require.Equal(t, "", got.GetImageName())
	require.Equal(t, "serve --port 8080", got.GetCommand())
	require.False(t, got.GetUseImagePorts())
	require.True(t, got.GetAutoUpgrade())
	require.True(t, got.GetIsDeclarativeDeploy())
	require.Equal(t, "production", got.GetEnvironment())

	require.Equal(t, map[string]string{"LOG_LEVEL": "info"}, got.GetEnv())
	require.Equal(t, map[string]string{
		"team":              "platform",
		"verv.tag.backend":  "true",
		"verv.tag.critical": "true",
	}, got.GetLabels())

	settings := got.GetSettings()
	require.NotNil(t, settings)
	require.Len(t, settings.GetPorts(), 2)
	require.Equal(t, uint32(8080), settings.GetPorts()[0].GetServicePortNumber())
	require.Equal(t, velez_api.Port_tcp, settings.GetPorts()[0].GetProtocol())
	require.Equal(t, uint32(80), settings.GetPorts()[0].GetExposedTo())
	require.Equal(t, uint32(9090), settings.GetPorts()[1].GetServicePortNumber())
	require.Equal(t, velez_api.Port_udp, settings.GetPorts()[1].GetProtocol())
	require.Nil(t, settings.GetPorts()[1].ExposedTo)
	require.Len(t, settings.GetVolumes(), 1)
	require.Equal(t, "data", settings.GetVolumes()[0].GetVolumeName())
	require.Equal(t, "/var/lib/data", settings.GetVolumes()[0].GetContainerPath())

	hc := got.GetHealthcheck()
	require.NotNil(t, hc)
	require.Equal(t, "curl -f localhost:8080/health", hc.GetCommand())
	require.Equal(t, uint32(10), hc.GetIntervalSecond())
	require.Equal(t, uint32(5), hc.GetTimeoutSecond())
	require.Equal(t, uint32(3), hc.GetRetries())

	restart := got.GetRestart()
	require.NotNil(t, restart)
	require.Equal(t, velez_api.RestartPolicyType_on_failure, restart.GetType())
	require.Equal(t, uint32(5), restart.GetFailureCount())

	hardware := got.GetHardware()
	require.NotNil(t, hardware)
	require.Equal(t, float32(0.5), hardware.GetCpu())
	require.Equal(t, uint32(512), hardware.GetRamMb())
}

func TestBoxResolver_ResolveRequest_DeployRequestImageAlwaysWins(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	cases := []struct {
		name          string
		descriptorImg string
		requestImg    string
		wantImage     string
	}{
		{
			name:          "request image wins when both set",
			descriptorImg: testDescriptorImage,
			requestImg:    "registry/example:from-request",
			wantImage:     "registry/example:from-request",
		},
		{
			name:          "descriptor image used when request omits it",
			descriptorImg: testDescriptorImage,
			requestImg:    "",
			wantImage:     testDescriptorImage,
		},
		{
			name:          "empty when neither set",
			descriptorImg: "",
			requestImg:    "",
			wantImage:     "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			descriptor := verv.Descriptor{
				Index: verv.Index{Service: verv.Service{Name: testSvcName}},
				Deployment: verv.Deployment{
					App: verv.App{Image: tc.descriptorImg},
				},
			}

			got, err := resolver.ResolveRequest(context.Background(), descriptor, "dev", tc.requestImg)
			require.NoError(t, err)
			require.Equal(t, tc.wantImage, got.GetImageName())
		})
	}
}

func TestBoxResolver_ResolveRequest_BoxResolvesToHardware(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	cases := []struct {
		name         string
		app          verv.App
		rootBox      string
		wantHardware *velez_api.Container_Hardware
	}{
		{
			name:    "own box wins over root box",
			app:     verv.App{Box: boxSmall},
			rootBox: boxMedium,
			wantHardware: &velez_api.Container_Hardware{
				Cpu:   ptrFloat32(0.5),
				RamMb: ptrUint32(512),
			},
		},
		{
			name:    "exact resources block wins over box",
			app:     verv.App{Box: boxSmall, Resources: verv.Sizing{Cpu: 4, RamMb: 4096}},
			rootBox: boxMedium,
			wantHardware: &velez_api.Container_Hardware{
				Cpu:   ptrFloat32(4),
				RamMb: ptrUint32(4096),
			},
		},
		{
			name:         "no box anywhere leaves hardware unset",
			app:          verv.App{},
			rootBox:      "",
			wantHardware: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			descriptor := verv.Descriptor{
				Index: verv.Index{
					Service: verv.Service{Name: testSvcName},
					Box:     tc.rootBox,
				},
				Deployment: verv.Deployment{App: tc.app},
			}

			got, err := resolver.ResolveRequest(context.Background(), descriptor, "dev", "")
			require.NoError(t, err)
			require.Equal(t, tc.wantHardware, got.GetHardware())
		})
	}
}

func TestBoxResolver_ResolveRequest_UnknownBoxIsAnError(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	descriptor := verv.Descriptor{
		Index: verv.Index{
			Service: verv.Service{Name: testSvcName},
			Box:     "gigantic",
		},
	}

	_, err := resolver.ResolveRequest(context.Background(), descriptor, "dev", "")
	require.Error(t, err)
}

// TestBoxResolver_ResolveRequest_JsonRoundTrip guards the resolved
// CreateSmerd.Request against the encoding/json-vs-oneof hazard documented in
// CLAUDE.md: a proto message with a oneof interface field marshals fine but
// silently fails to unmarshal back into a concrete type. As of this wave
// CreateSmerd_Request carries no oneof field (verv/plain were de-oneof'd -
// see the comment above CreateSmerd_Request.Verv in velez_api.pb.go) so this
// currently just confirms plain encoding/json round-trips every mapped
// field; it starts failing the moment a oneof is reintroduced, which is
// exactly the point.
func TestBoxResolver_ResolveRequest_JsonRoundTrip(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	descriptor := fullDescriptor()

	want, err := resolver.ResolveRequest(context.Background(), descriptor, "production", "registry/example:v2")
	require.NoError(t, err)

	raw, err := json.Marshal(want)
	require.NoError(t, err)

	got := &velez_api.CreateSmerd_Request{}

	err = json.Unmarshal(raw, got)
	require.NoError(t, err)

	require.Equal(t, want.GetName(), got.GetName())
	require.Equal(t, want.GetImageName(), got.GetImageName())
	require.Equal(t, want.GetRepo(), got.GetRepo())
	require.Equal(t, want.GetCommand(), got.GetCommand())
	require.Equal(t, want.GetEnv(), got.GetEnv())
	require.Equal(t, want.GetLabels(), got.GetLabels())
	require.Equal(t, want.GetUseImagePorts(), got.GetUseImagePorts())
	require.Equal(t, want.GetAutoUpgrade(), got.GetAutoUpgrade())
	require.Equal(t, want.GetIsDeclarativeDeploy(), got.GetIsDeclarativeDeploy())
	require.Equal(t, want.GetEnvironment(), got.GetEnvironment())
	require.NotNil(t, got.GetHealthcheck())
	require.Equal(t, want.GetHealthcheck().GetCommand(), got.GetHealthcheck().GetCommand())
	require.NotNil(t, got.GetRestart())
	require.Equal(t, want.GetRestart().GetType(), got.GetRestart().GetType())
	require.Equal(t, want.GetRestart().GetFailureCount(), got.GetRestart().GetFailureCount())
	require.NotNil(t, got.GetHardware())
	require.Equal(t, want.GetHardware().GetCpu(), got.GetHardware().GetCpu())
	require.NotNil(t, got.GetSettings())
	require.Len(t, got.GetSettings().GetPorts(), len(want.GetSettings().GetPorts()))
}

func ptrFloat32(v float32) *float32 {
	return &v
}

func ptrUint32(v uint32) *uint32 {
	return &v
}
