package domain

const (
	RunnerBuildkitAlias = "buildkit"
	RunnerBuildkitPort  = 1234

	runnerBuildkitSuffix            = "-buildkit"
	runnerBuildkitStateVolumeSuffix = "-buildkit-state"
)

// RunnerBuildkitServiceName is the BuildKit sidecar container name; the
// runner's private BuildKit network carries the same name.
func RunnerBuildkitServiceName(runnerName string) string {
	return runnerName + runnerBuildkitSuffix
}

func RunnerBuildkitStateVolumeName(runnerName string) string {
	return runnerName + runnerBuildkitStateVolumeSuffix
}
