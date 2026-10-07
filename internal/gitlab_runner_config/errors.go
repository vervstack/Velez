package gitlab_runner_config

import (
	"go.redsock.ru/rerrors"
)

var (
	ErrRunnerEntryMissing = rerrors.New("gitlab-runner config.toml has no [[runners]] entry")

	ErrRunnerEnvironmentInvalid = rerrors.New(
		"gitlab-runner config.toml [[runners]] environment is not an array of strings")
)
