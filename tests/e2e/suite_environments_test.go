package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

// EnvironmentsSuite covers multi-environment behavior against real Docker on a
// single-node (local_storage) Velez - no matreshka, no postgres, so it never
// touches the shared WithMatreshka() fixture.
//
// Environments are seeded straight from config here: local_storage builds its
// environments storage with environments.NewStatic(cfg.Environment.Environments,
// cfg.Environment.ContainerSuffix), i.e. one PROD row carrying the node's
// ContainerSuffix plus one row per configured name (suffix = name). That's what
// WithContainerSuffix + WithEnvironments below drive.
//
// Two of the four tests are deliberately RED and skipped - see their
// t.Skip reasons; they describe isolation the current implementation does not
// provide yet (the environment suffix is only a container LABEL, never part of
// the container NAME, and DropSmerd ignores its environment field entirely).
type EnvironmentsSuite struct {
	suite.Suite
}

const (
	// e2eDefaultSuffix is deliberately non-empty: with the default (empty)
	// ContainerSuffix, "scoped to PROD" and "not scoped at all" produce the
	// same Docker filter, and the assertions below couldn't tell the default
	// fallback from no scoping at all.
	e2eDefaultSuffix = "e2eprod"
	e2eStageEnv      = "E2ESTAGE"

	// Container names double as Docker hostnames, which are capped at 64
	// characters - GetServiceName(t) on a suite subtest blows past that, so
	// these tests use their own short, suite-unique names.
	e2eEnvDefaultName = "e2e_env_default"
	e2eEnvProdName    = "e2e_env_prod"
	e2eEnvStageName   = "e2e_env_stage"
	e2eEnvSharedName  = "e2e_env_shared"
)

// Backward compatibility: a request that carries NO environment at all - every
// caller that predates the feature, including the rest of this e2e suite - must
// silently land in the default environment (PROD) and be stamped with its
// suffix, not with an empty one.
func (s *EnvironmentsSuite) Test_EmptyEnvironment_UsesDefaultSuffix() {
	t := s.T()

	serviceName := e2eEnvDefaultName
	env := NewEnvironment(t, WithContainerSuffix(e2eDefaultSuffix))

	createReq := &velez_api.CreateSmerd_Request{
		Name:         serviceName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	created := env.CreateSmerd(t, createReq)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())
	require.Equal(t, e2eDefaultSuffix, created.GetLabels()[labels.SuffixLabel],
		"an environment-less create must be stamped with the default environment's suffix")

	// The same request spelled with an explicit PROD must find that container,
	// proving "" and PROD are the same environment.
	listReq := &velez_api.ListSmerds_Request{
		Environment: environments.DefaultEnvironmentName,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	listed := env.ListSmerds(t, t.Context(), listReq)
	require.Len(t, listed.GetSmerds(), 1)
	require.Equal(t, serviceName, listed.GetSmerds()[0].GetName())

	// ... and so must an environment-less list.
	unscopedReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{testCaseNameLabel: t.Name()},
	}

	unscoped := env.ListSmerds(t, t.Context(), unscopedReq)
	require.Len(t, unscoped.GetSmerds(), 1)
}

// Two differently-named containers in two different environments must get two
// different suffixes, and a list scoped to one environment must not see the
// other's container.
func (s *EnvironmentsSuite) Test_TwoEnvironments_AreListScoped() {
	t := s.T()

	env := NewEnvironment(t,
		WithContainerSuffix(e2eDefaultSuffix),
		WithEnvironments([]string{e2eStageEnv}))

	prodReq := &velez_api.CreateSmerd_Request{
		Name:         e2eEnvProdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	prodSmerd := env.CreateSmerd(t, prodReq)
	require.Equal(t, e2eDefaultSuffix, prodSmerd.GetLabels()[labels.SuffixLabel])

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         e2eEnvStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  e2eStageEnv,
	}

	stageSmerd := env.CreateSmerd(t, stageReq)
	require.Equal(t, e2eStageEnv, stageSmerd.GetLabels()[labels.SuffixLabel])

	require.NotEqual(t,
		prodSmerd.GetLabels()[labels.SuffixLabel],
		stageSmerd.GetLabels()[labels.SuffixLabel],
		"environments must not share a suffix")

	stageListReq := &velez_api.ListSmerds_Request{
		Environment: e2eStageEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	stageList := env.ListSmerds(t, t.Context(), stageListReq)
	require.Len(t, stageList.GetSmerds(), 1)
	require.Equal(t, e2eEnvStageName, stageList.GetSmerds()[0].GetName())

	prodListReq := &velez_api.ListSmerds_Request{
		Environment: environments.DefaultEnvironmentName,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	prodList := env.ListSmerds(t, t.Context(), prodListReq)
	require.Len(t, prodList.GetSmerds(), 1)
	require.Equal(t, e2eEnvProdName, prodList.GetSmerds()[0].GetName())
}

// RED. The headline multi-environment promise: the SAME logical service name
// deployed into two environments must yield two distinct containers.
//
// It cannot pass today, for two stacked reasons - both verified by running this
// test with the Skip removed (observed: the second CreateSmerd returns the very
// same container UUID as the first):
//
//  1. Tasks are keyed by (entity_id, action) with no environment component
//     (internal/transport/velez_api_impl/smerd_create.go enqueues on
//     req.GetName()), so the second create dedups onto the first environment's
//     already-DONE create_smerd task and never runs.
//  2. Even without that, the resolved suffix is written only as the
//     labels.SuffixLabel container LABEL
//     (internal/clients/node_clients/docker/client.go ContainerCreate) - the
//     Docker container NAME stays the bare req.GetName(), which is globally
//     unique per daemon.
func (s *EnvironmentsSuite) Test_SameNameInTwoEnvironments_AreDistinctContainers() {
	t := s.T()

	t.Skip("needs: (1) the environment folded into the jobs-engine entity id so " +
		"two environments don't dedup onto one create_smerd task - see " +
		"internal/transport/velez_api_impl/smerd_create.go Enqueue(req.GetName(), ...); " +
		"and (2) the environment suffix applied to the Docker container NAME, not just " +
		"the VELEZ_SUFFIX label - see internal/clients/node_clients/docker/client.go " +
		"ContainerCreate. Observed today: the second create returns the first " +
		"environment's container.")

	env := NewEnvironment(t,
		WithContainerSuffix(e2eDefaultSuffix),
		WithEnvironments([]string{e2eStageEnv}))

	prodReq := &velez_api.CreateSmerd_Request{
		Name:         e2eEnvSharedName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	prodSmerd := env.CreateSmerd(t, prodReq)

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         e2eEnvSharedName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  e2eStageEnv,
	}

	stageSmerd := env.CreateSmerd(t, stageReq)

	require.NotEqual(t, prodSmerd.GetUuid(), stageSmerd.GetUuid(),
		"same name in two environments must be two distinct containers")
	require.Equal(t, e2eDefaultSuffix, prodSmerd.GetLabels()[labels.SuffixLabel])
	require.Equal(t, e2eStageEnv, stageSmerd.GetLabels()[labels.SuffixLabel])

	stageListReq := &velez_api.ListSmerds_Request{
		Environment: e2eStageEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	stageList := env.ListSmerds(t, t.Context(), stageListReq)
	require.Len(t, stageList.GetSmerds(), 1)
	require.Equal(t, stageSmerd.GetUuid(), stageList.GetSmerds()[0].GetUuid())
}

// RED. A DropSmerd scoped to one environment must not touch another
// environment's container.
//
// It cannot pass today: DropSmerd's `environment` field is validated at the
// transport layer and then dropped on the floor - internal/jobs/drop_smerd.go
// builds its worklist from req.GetUuids()+req.GetName() and removes by
// name/uuid, with no suffix/label filter anywhere. So dropping "X in PROD"
// happily removes X even though X only exists in STAGE (verified by running
// this test with the Skip removed: the STAGE container is gone afterwards).
func (s *EnvironmentsSuite) Test_DropSmerd_ScopedToEnvironment_LeavesOtherEnvironmentAlone() {
	t := s.T()

	t.Skip("needs: DropSmerd to filter its worklist by the resolved environment suffix - " +
		"internal/jobs/drop_smerd.go currently removes purely by name/uuid and ignores " +
		"DropSmerd_Request.environment entirely.")

	env := NewEnvironment(t,
		WithContainerSuffix(e2eDefaultSuffix),
		WithEnvironments([]string{e2eStageEnv}))

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         e2eEnvStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  e2eStageEnv,
	}

	stageSmerd := env.CreateSmerd(t, stageReq)

	// Drop the STAGE container's name, but scoped to PROD: nothing must happen.
	dropReq := &velez_api.DropSmerd_Request{
		Name:        []string{e2eEnvStageName},
		Environment: environments.DefaultEnvironmentName,
	}

	env.DropSmerd(t.Context(), t, dropReq)

	stageListReq := &velez_api.ListSmerds_Request{
		Environment: e2eStageEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	stageList := env.ListSmerds(t, t.Context(), stageListReq)
	require.Len(t, stageList.GetSmerds(), 1,
		"a PROD-scoped drop must not remove a STAGE container")
	require.Equal(t, stageSmerd.GetUuid(), stageList.GetSmerds()[0].GetUuid())
}

func Test_Environments(t *testing.T) {
	suite.Run(t, new(EnvironmentsSuite))
}
