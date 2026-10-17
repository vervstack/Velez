package e2e

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/tests/dind"
)

func TestMain(m *testing.M) {
	flag.Parse()

	// -test.list only prints names; it must not boot a daemon (tests/e2egroups
	// lists tests to verify the group partition).
	if isListOnly() {
		os.Exit(m.Run())
	}

	os.Exit(runSuite(m))
}

func isListOnly() bool {
	listFlag := flag.Lookup("test.list")

	return listFlag != nil && listFlag.Value.String() != ""
}

// runSuite owns the DinD lifecycle for the whole package: bring up one
// disposable daemon, point DOCKER_HOST at it for every test, run the suite,
// then tear everything down. Split out of TestMain so deferred cleanup runs
// before os.Exit.
func runSuite(m *testing.M) int {
	ctx := context.Background()

	env, err := dind.Setup(ctx, dind.Options{})
	if err != nil {
		log.Error().Err(err).Msg("error setting up dind harness")

		return 1
	}

	sharedDind = env

	defer teardown(env)

	// ctrl-C / CI cancellation still removes the container.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		teardown(env)

		os.Exit(1)
	}()

	err = os.Setenv("DOCKER_HOST", env.DockerHost)
	if err != nil {
		log.Error().Err(err).Msg("error setting DOCKER_HOST")

		return 1
	}

	err = os.Setenv(dind.EnvActive, "1")
	if err != nil {
		log.Error().Err(err).Msg("error setting dind marker env")

		return 1
	}

	err = env.Reset(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error resetting dind state")

		return 1
	}

	err = env.Seed(ctx, seedImages()...)
	if err != nil {
		log.Error().Err(err).Msg("error seeding dind images")

		return 1
	}

	err = env.EnsureNetwork(ctx, dindEnsureNetworks...)
	if err != nil {
		log.Error().Err(err).Msg("error ensuring dind networks")

		return 1
	}

	if os.Getenv(dind.EnvPrepareOnly) != "" {
		return 0
	}

	code := m.Run()

	if sharedMatreshka != nil {
		stopErr := sharedMatreshka.Stop()
		if stopErr != nil {
			log.Error().Err(stopErr).Msg("error stopping shared matreshka instance")
		}
	}

	stopSharedHeadscale()

	return code
}

func teardown(env *dind.Env) {
	err := env.Teardown()
	if err != nil {
		log.Error().Err(err).Msg("error tearing down dind harness")
	}
}
