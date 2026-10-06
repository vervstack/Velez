package gitlab_runner_config

import (
	"go.redsock.ru/rerrors"
)

var (
	ErrRunnerEntryMissing = rerrors.New("gitlab-runner config.toml has no [[runners]] entry")
)
