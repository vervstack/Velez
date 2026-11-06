package selfswap

import (
	"time"
)

const (
	// SwapperContainerName is the name of the throwaway container Velez
	// creates to carry out its own redeploy. It runs the same image as the
	// target Velez container with VELEZ_SELFSWAP=1 and performs the
	// stop/rename/recreate swap against the target.
	SwapperContainerName = "velez-swapper"
	// RollbackContainerName is the name the target Velez container is renamed
	// to for the rollback window before it is dropped for good.
	RollbackContainerName = "velez-rollback"

	// EnvSelfSwap marks a container (or the bare binary's environment) as the
	// swapper mode: main() branches on it before building the regular app.
	EnvSelfSwap = "VELEZ_SELFSWAP"
	// EnvTargetContainerId carries the id of the Velez container to swap out.
	EnvTargetContainerId = "VELEZ_SELFSWAP_TARGET_CONTAINER_ID"
	// EnvImage carries the image the new Velez container is recreated with.
	// Empty means the target's own image (re-pull).
	EnvImage = "VELEZ_SELFSWAP_IMAGE"
	// EnvHealthTimeout bounds how long the swapper waits for the new Velez
	// container to become healthy before rolling back.
	EnvHealthTimeout = "VELEZ_SELFSWAP_HEALTH_TIMEOUT"
	// EnvRollbackWindow bounds how long the renamed old container
	// (velez-rollback) is kept around for rollback before being dropped.
	EnvRollbackWindow = "VELEZ_SELFSWAP_ROLLBACK_WINDOW"

	// DefaultHealthTimeout is used when EnvHealthTimeout is unset.
	DefaultHealthTimeout = time.Second * 60
	// DefaultRollbackWindow is used when EnvRollbackWindow is unset.
	DefaultRollbackWindow = time.Minute * 10
)
