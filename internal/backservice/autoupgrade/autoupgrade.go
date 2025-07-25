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
	"go.redsock.ru/rerrors"
	"golang.org/x/sync/errgroup"

	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type AutoUpgrade struct {
	dockerAPI client.APIClient

	starter sync.Once
	closer  sync.Once

	stopC chan struct{}
	errC  chan error

	checkPeriod time.Duration

	pipeliner pipelines.Pipeliner
}

func New(api client.APIClient, checkPeriod time.Duration, pipeliner pipelines.Pipeliner) *AutoUpgrade {
	return &AutoUpgrade{
		dockerAPI: api,
		stopC:     make(chan struct{}),
		errC:      make(chan error),

		checkPeriod: max(checkPeriod, time.Second*30),
		pipeliner:   pipeliner,
	}
}

func (au *AutoUpgrade) Start() error {

	go au.starter.Do(func() {
		err := au.do()
		if err != nil {
			//	 TODO log
		}

		for {
			select {
			case <-time.After(au.checkPeriod):
				err = au.do()
				if err != nil {
					//	 TODO log
				}
			case <-au.stopC:
			}

		}
	})

	return nil
}

func (au *AutoUpgrade) Stop() error {
	au.closer.Do(func() {
		close(au.stopC)
	})

	return nil
}

func (au *AutoUpgrade) do() error {
	ctx := context.Background()

	smerds, err := au.getAutoUpdateSmerds(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error getting smerds to upgrade")
	}

	eg := errgroup.Group{}

	for _, smerd := range smerds {
		var newImage *string
		newImage, err = au.getNewImageVersion(ctx, smerd.Image)
		if err != nil {
			// TODO log
			continue
		}
		if newImage == nil {
			continue
		}

		r := domain.UpgradeSmerd{
			Name:  smerd.Names[0][1:],
			Image: *newImage,
		}

		eg.Go(func() error {
			runner := au.pipeliner.UpgradeSmerd(r)
			err = runner.Run(ctx)
			if err != nil {
				return rerrors.Wrapf(err, "error upgrading smerd %s", r.Name)
			}

			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return rerrors.Wrap(err)
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

	return &newImage, nil
}

func imageNameWithoutTag(in string) string {
	tagIndx := strings.Index(in, ":")
	if tagIndx != -1 {
		in = in[:tagIndx]
	}

	return in
}
