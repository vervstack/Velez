package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
)

const (
	testCustomLabel = "custom"
)

func newLabeledTestContainer(
	environment, name string, isRunning bool, containerLabels map[string]string,
) labeledContainer {
	return labeledContainer{
		environment: environment,
		info: container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{
				Name:       "/" + name,
				State:      &container.State{Running: isRunning},
				HostConfig: &container.HostConfig{},
			},
			Config: &container.Config{Image: "img", Labels: containerLabels},
		},
	}
}

func Test_GroupLabeledServices_SkipsUnlabeledAndSidecars(t *testing.T) {
	containers := []labeledContainer{
		newLabeledTestContainer("prod", "plain", true, map[string]string{}),
		newLabeledTestContainer("prod", "side", true, map[string]string{
			labels.VervServiceLabel: testSvc,
			labels.Sidecar:          labelTrueValue,
		}),
		newLabeledTestContainer("prod", testSvc, true, map[string]string{labels.VervServiceLabel: testSvc}),
	}

	services := groupLabeledServices(containers)

	require.Len(t, services, 1)
	require.Equal(t, testSvc, services[0].name)
	require.Equal(t, testSvc, services[0].primary.containerName())
}

func Test_GroupLabeledServices_GroupsByServiceAndPrefersRunning(t *testing.T) {
	containers := []labeledContainer{
		newLabeledTestContainer("prod", "a_old", false, map[string]string{labels.VervServiceLabel: "a"}),
		newLabeledTestContainer("prod", "b", true, map[string]string{labels.VervServiceLabel: "b"}),
		newLabeledTestContainer("prod", "a", true, map[string]string{labels.VervServiceLabel: "a"}),
	}

	services := groupLabeledServices(containers)

	require.Len(t, services, 2)
	require.Equal(t, "a", services[0].name)
	require.Equal(t, "a", services[0].primary.containerName())
	require.Equal(t, "b", services[1].name)
}

func Test_GroupLabeledServices_DisplayNameFallsBackToName(t *testing.T) {
	containers := []labeledContainer{
		newLabeledTestContainer("prod", "gitlab_runner_Artel", true, map[string]string{
			labels.VervServiceLabel: "gitlab_runner_Artel",
			labels.DisplayNameLabel: "Artel",
		}),
		newLabeledTestContainer("prod", "plain", true, map[string]string{labels.VervServiceLabel: "plain"}),
	}

	services := groupLabeledServices(containers)

	require.Equal(t, "Artel", services[0].displayName)
	require.Equal(t, "plain", services[1].displayName)
}

func Test_GroupLabeledServices_PicksAasContainers(t *testing.T) {
	containers := []labeledContainer{
		newLabeledTestContainer("prod", "pgaas_db", true, map[string]string{
			labels.VervServiceLabel:   "pgaas_db",
			labels.PgaasInstanceLabel: labelTrueValue,
		}),
		newLabeledTestContainer("prod", "gitlab_runner_r", true, map[string]string{
			labels.VervServiceLabel:    "gitlab_runner_r",
			labels.RunnerInstanceLabel: labelTrueValue,
		}),
		newLabeledTestContainer("prod", "cr_reg", true, map[string]string{
			labels.VervServiceLabel:         "cr_reg",
			labels.RegistryaasInstanceLabel: labelTrueValue,
		}),
	}

	services := groupLabeledServices(containers)

	require.Len(t, services, 3)
	require.NotNil(t, services[0].pg)
	require.Nil(t, services[0].runner)
	require.NotNil(t, services[1].runner)
	require.NotNil(t, services[2].registry)
}

func Test_InspectedContainerSpec_PreservesLabelsAndSetsEnvironment(t *testing.T) {
	c := newLabeledTestContainer("staging", testSvc, true, map[string]string{
		labels.VervServiceLabel: testSvc,
		labels.DisplayNameLabel: "Service",
		testCustomLabel:         "value",
	})

	spec := inspectedContainerSpec(c.containerName(), c.environment, map[string]string{}, c.info)

	require.Equal(t, testSvc, spec.GetName())
	require.Equal(t, "staging", spec.GetEnvironment())
	require.Equal(t, "img", spec.GetImageName())
	require.Equal(t, "Service", spec.GetLabels()[labels.DisplayNameLabel])
	require.Equal(t, "value", spec.GetLabels()[testCustomLabel])
}

func Test_RegisteredContainerSpec_OverlaysServiceLabels(t *testing.T) {
	c := newLabeledTestContainer("prod", testSvc, true, map[string]string{testCustomLabel: "value"})

	spec := registeredContainerSpec(c.containerName(), c.environment, testSvc, c.info)

	require.Equal(t, testSvc, spec.GetLabels()[labels.VervServiceLabel])
	require.Equal(t, testSvc, spec.GetLabels()[labels.DisplayNameLabel])
	require.Equal(t, "value", spec.GetLabels()[testCustomLabel])
}

func Test_RegistryInstanceSecretRef_MatchesCreatePath(t *testing.T) {
	require.Equal(t, registryInstanceSecretRef("cr_reg"), container_derived.RegistryInstanceSecretRef("cr_reg"))
}

func Test_PgInstanceSecretRef_MatchesRegisterPath(t *testing.T) {
	require.Equal(t, registeredPgSecretRef("pgaas_db"), container_derived.PgInstanceSecretRef("pgaas_db"))
}
