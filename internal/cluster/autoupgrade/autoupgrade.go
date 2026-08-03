package autoupgrade

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"golang.org/x/sync/errgroup"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	defaultCheckPeriod = time.Second * 30

	// taskWatchTimeout bounds how long one auto-upgrade blocks waiting for its
	// upgrade_smerd task to reach a terminal status - same safety-net role and
	// same size as velez_api_impl's upgradeSmerdWatchTimeout.
	taskWatchTimeout = time.Second * 120
)

// taskRunner is the slice of jobs.Engine this worker needs: enqueue a task
// and follow it to a terminal status.
type taskRunner interface {
	Enqueue(ctx context.Context, entityID, action string, initialContext any) (tasks_queries.VelezTask, error)
	Watch(ctx context.Context, entityID, action string) <-chan tasks_queries.VelezTask
}

type AutoUpgrade struct {
	dockerAPI client.APIClient

	starter sync.Once
	closer  sync.Once

	stopC chan struct{}

	checkPeriod time.Duration

	// jobsEngine replaces the deleted internal/pipelines.Pipeliner: each
	// upgrade is enqueued as a durable upgrade_smerd task and awaited
	// synchronously, mirroring velez_api_impl/smerd_upgrade.go's facade.
	jobsEngine taskRunner
}

func New(api client.APIClient, checkPeriod time.Duration, jobsEngine jobs.Engine) *AutoUpgrade {
	return &AutoUpgrade{
		dockerAPI: api,
		stopC:     make(chan struct{}),

		checkPeriod: max(checkPeriod, defaultCheckPeriod),
		jobsEngine:  jobsEngine,
	}
}

func (au *AutoUpgrade) Start(ctx context.Context) error {
	go au.starter.Do(func() {
		err := au.do(ctx)
		if err != nil {
			log.Err(err).
				Msg("autoupgrade start failed")
		}

		for {
			select {
			case <-time.After(au.checkPeriod):
				err = au.do(ctx)
				if err != nil {
					log.Err(err).
						Msg("autoupgrade failed")
				}
			case <-au.stopC:
				return
			case <-ctx.Done():
				return
			}
		}
	})

	return nil
}

func (au *AutoUpgrade) Stop() error {
	au.closer.Do(
		func() {
			close(au.stopC)
		})

	return nil
}

func (au *AutoUpgrade) do(ctx context.Context) error {
	smerds, err := au.getAutoUpdateSmerds(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error getting smerds to upgrade")
	}

	eg := errgroup.Group{}

	for _, smerd := range smerds {
		var newImage *string

		newImage, err = au.getNewImageVersion(ctx, smerd.Image)
		if err != nil {
			log.Err(err).
				Str("image", smerd.Image).
				Msg("error getting new image version")

			continue
		}

		if newImage == nil {
			continue
		}

		// Environment is deliberately left empty, which resolves to the
		// default (PROD) environment.
		//
		// KNOWN LIMITATION, carried over unchanged from the pipeliner this
		// replaced: auto-upgrade has never been environment-aware. The
		// container.Summary rows it works from carry no environment name, so
		// a smerd running in a non-default environment is looked up - and
		// upgraded - as if it lived in the default one. Teaching this worker
		// to reverse-resolve an environment from labels.SuffixLabel is a
		// separate feature, out of scope for deleting internal/pipelines.
		upgradeReq := &velez_api.UpgradeSmerd_Request{
			Name:  smerd.Names[0][1:],
			Image: *newImage,
		}

		eg.Go(func() error {
			upgradeErr := au.upgrade(ctx, upgradeReq)
			if upgradeErr != nil {
				return rerrors.Wrapf(upgradeErr, "error upgrading smerd %s", upgradeReq.GetName())
			}

			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return rerrors.Wrap(err, "error returned from error group")
	}

	return nil
}

// upgrade enqueues an upgrade_smerd task and blocks until it reaches a
// terminal status.
func (au *AutoUpgrade) upgrade(ctx context.Context, req *velez_api.UpgradeSmerd_Request) error {
	initialContext := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: req,
	}

	_, err := au.jobsEngine.Enqueue(ctx, req.GetName(), jobs.UpgradeSmerdAction, initialContext)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing upgrade_smerd task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, taskWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range au.jobsEngine.Watch(watchCtx, req.GetName(), jobs.UpgradeSmerdAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for upgrade_smerd task, last status: %q",
			finalTask.Status,
		)
	}

	if isFailed {
		return rerrors.New(finalTask.Error.String)
	}

	return nil
}

func (au *AutoUpgrade) getAutoUpdateSmerds(ctx context.Context) ([]container.Summary, error) {
	listReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			labels.AutoUpgrade: "true",
		},
	}

	conts, err := dockerutils.ListContainers(ctx, au.dockerAPI, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	return conts, nil
}

func (au *AutoUpgrade) getNewImageVersion(ctx context.Context, imageName string) (*string, error) {
	imageBase := imageNameWithoutTag(imageName)

	repo, err := name.NewRepository(imageBase)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting image base repo")
	}

	tags, err := remote.List(repo,
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing tags")
	}

	newImage := imageBase + ":" + tags[len(tags)-1]
	if newImage == imageName {
		return nil, nil
	}

	return &newImage, nil
}

func imageNameWithoutTag(in string) string {
	tagIndx := strings.Index(in, ":")
	if tagIndx != -1 {
		in = in[:tagIndx]
	}

	return in
}
