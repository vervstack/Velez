package e2e

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/tests/dind"
)

func TestMain(m *testing.M) {
	os.Exit(runSuite(m))
}

// runSuite owns the DinD lifecycle for the whole package: bring up one
// disposable daemon, point DOCKER_HOST at it for every test, run the suite,
// then tear everything down. Split out of TestMain so deferred cleanup runs
// before os.Exit.
func runSuite(m *testing.M) int {
	ctx := context.Background()

	opts := dind.Options{
		Publish: dindPublishPorts(),
	}

	env, err := dind.Setup(ctx, opts)
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

	err = env.Seed(ctx, dindSeedImages...)
	if err != nil {
		log.Error().Err(err).Msg("error seeding dind images")

		return 1
	}

	err = env.EnsureNetwork(ctx, dindEnsureNetworks...)
	if err != nil {
		log.Error().Err(err).Msg("error ensuring dind networks")

		return 1
	}

	code := m.Run()

	if sharedMatreshka != nil {
		stopErr := sharedMatreshka.Stop()
		if stopErr != nil {
			log.Error().Err(stopErr).Msg("error stopping shared matreshka instance")
		}
	}

	return code
}

func teardown(env *dind.Env) {
	err := env.Teardown()
	if err != nil {
		log.Error().Err(err).Msg("error tearing down dind harness")
	}
}
