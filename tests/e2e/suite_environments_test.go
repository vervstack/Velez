//go:build e2e_full

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
// Test_SameNameInTwoEnvironments_AreDistinctContainers used to be t.Skip'd on
// the jobs-engine dedup bug (tasks keyed on the bare smerd name, so two
// environments' same-named creates collided on velez.tasks' UNIQUE
// (entity_id, action)); it is now GREEN - smerd_create.go folds the resolved
// environment suffix into the entity id via jobs.SmerdEntityID. The two
// DropSmerd tests below used to be deliberately RED,
// documenting bugs in internal/jobs/drop_smerd.go's ContainerRuntime.Remove
// wiring (see docs/container_runtimes/roadmap.md); both are now GREEN -
// labelBasedRuntime.Remove resolves bare/suffixed/UUID identifiers to a real
// container and checks its labels.SuffixLabel against r.suffix before
// removing, closing both the bare-name-no-op and the UUID-cross-environment
// bugs.
type EnvironmentsSuite struct {
	suite.Suite
}

const (
	// e2eDefaultSuffix is deliberately non-empty: with an empty
	// ContainerSuffix, the suffixed and bare forms of a container name are
	// the same string, and the assertions below couldn't tell whether
	// suffix-handling is actually exercised.
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
	env := Planes[0].NewEnvironment(t, WithContainerSuffix(e2eDefaultSuffix))

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
	// The Docker container name carries the environment's suffix (see
	// expectedContainerName in suite_container_runtime_test.go), but Smerd.Name
	// is always the virtual/logical name - ContainerRuntime.ListContainers
	// strips the suffix back off before ListSmerds ever sees it (see
	// docs/container_runtimes/interface_design.md).
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

	env := Planes[0].NewEnvironment(t,
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

// The headline multi-environment promise: the SAME logical service name
// deployed into two environments must yield two distinct containers.
//
// It used to fail for two stacked reasons, both now fixed:
//
//  1. FIXED (Phase 6, card #127). Tasks were keyed by (entity_id, action)
//     with no environment component (smerd_create.go enqueued on
//     req.GetName()), so the second create deduped onto the first
//     environment's already-DONE create_smerd task and never ran.
//     smerd_create.go now composes the entity id as jobs.SmerdEntityID(
//     resolvedSuffix, name).
//  2. FIXED (docs/container_runtimes Phase 1). The resolved suffix used to be
//     written only as the labels.SuffixLabel container LABEL, leaving the
//     Docker container NAME the bare req.GetName(). container_runtime's
//     labelBasedRuntime now derives the container name from it too.
func (s *EnvironmentsSuite) Test_SameNameInTwoEnvironments_AreDistinctContainers() {
	t := s.T()

	env := Planes[0].NewEnvironment(t,
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

// GREEN. Removing a container by its bare logical NAME must actually remove
// it, even in a non-default environment.
//
// This test replaces a previous version of this scenario
// (Test_DropSmerd_ScopedToEnvironment_LeavesOtherEnvironmentAlone, dropped
// "e2e_env_stage" scoped to PROD and asserted the STAGE container survived).
// That scenario was retired because it no longer proves what it claims to:
// since docs/container_runtimes Phase 1, the actual Docker container name for
// a non-empty-suffix environment is "<name>_<suffix>"
// (labelBasedRuntime.ContainerCreate), so a drop by the BARE name
// "e2e_env_stage" never matches ANY real container, in ANY environment -
// Docker's ContainerRemove returns "No such container", which
// dropContainerJob's idempotent semantics treat as success. The old test
// passed, but for the wrong reason: it observed a no-op, not correct
// environment scoping. (Verified empirically: temporarily un-skipping and
// running that old scenario alone now reports PASS with zero repro.)
//
// The REAL bug this test used to prove instead, now fixed: dropping by bare
// name used to be a silent no-op EVEN WHEN SCOPED TO THE CONTAINER'S OWN
// environment, because internal/jobs/drop_smerd.go's dropContainerJob never
// asked the resolved ContainerRuntime to translate the logical name into its
// suffixed Docker name before calling Remove. labelBasedRuntime.Remove now
// does that resolution itself (see its doc comment), so the drop below
// actually removes the container.
func (s *EnvironmentsSuite) Test_DropSmerd_ByBareName_SilentlyNoOpsInSuffixedEnvironment() {
	t := s.T()

	env := Planes[0].NewEnvironment(t,
		WithContainerSuffix(e2eDefaultSuffix),
		WithEnvironments([]string{e2eStageEnv}))

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         e2eEnvStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  e2eStageEnv,
	}

	env.CreateSmerd(t, stageReq)

	// Drop the STAGE container's bare logical name, correctly scoped to
	// STAGE - the environment it actually lives in.
	dropReq := &velez_api.DropSmerd_Request{
		Name:        []string{e2eEnvStageName},
		Environment: e2eStageEnv,
	}

	dropResp := env.DropSmerd(t.Context(), t, dropReq)
	require.Empty(t, dropResp.GetFailed(),
		"drop reports no failure - it thinks the removal succeeded")

	stageListReq := &velez_api.ListSmerds_Request{
		Environment: e2eStageEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	stageList := env.ListSmerds(t, t.Context(), stageListReq)
	require.Len(t, stageList.GetSmerds(), 0,
		"a same-environment bare-name drop must actually remove the container, "+
			"not silently no-op because the bare name never matched the suffixed Docker name")
}

// GREEN. Removing by UUID must still respect environment scoping: a drop
// resolved for one environment's runtime must not delete another
// environment's container, even though Docker's ContainerRemove matches
// purely on ID and a UUID can't be suffix-mangled the way a name can.
//
// This used to be the genuine cross-environment collision
// docs/container_runtimes warned about - unlike the bare-name scenario above
// (which just no-ops post name-suffixing), UUID-based removal is
// suffix-agnostic by construction: env.CreateSmerd's returned Uuid identifies
// one specific container on the shared daemon regardless of which
// environment's runtime resolves the Remove call. labelBasedRuntime.Remove
// now closes this: before removing anything, it inspects whichever container
// the identifier resolves to and compares its labels.SuffixLabel exactly
// against r.suffix (see its doc comment) - a UUID belonging to a different
// environment's suffix is treated as "not found here" and left untouched.
func (s *EnvironmentsSuite) Test_DropSmerd_ByUuid_CrossEnvironmentCollision() {
	t := s.T()

	env := Planes[0].NewEnvironment(t,
		WithContainerSuffix(e2eDefaultSuffix),
		WithEnvironments([]string{e2eStageEnv}))

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         e2eEnvStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  e2eStageEnv,
	}

	stageSmerd := env.CreateSmerd(t, stageReq)

	// Drop the STAGE container's UUID, but scoped to PROD: nothing must
	// happen to it - it belongs to STAGE, not PROD.
	dropReq := &velez_api.DropSmerd_Request{
		Uuids:       []string{stageSmerd.GetUuid()},
		Environment: environments.DefaultEnvironmentName,
	}

	env.DropSmerd(t.Context(), t, dropReq)

	stageListReq := &velez_api.ListSmerds_Request{
		Environment: e2eStageEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	stageList := env.ListSmerds(t, t.Context(), stageListReq)
	require.Len(t, stageList.GetSmerds(), 1,
		"a PROD-scoped drop must not remove a STAGE container, even by uuid")
	require.Equal(t, stageSmerd.GetUuid(), stageList.GetSmerds()[0].GetUuid())
}

func Test_Environments(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(EnvironmentsSuite))
}
