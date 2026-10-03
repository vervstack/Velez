package imagetags

import (
	"context"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"go.redsock.ru/rerrors"
	"golang.org/x/sync/errgroup"
)

const (
	maxRegistryTags   = 300
	registryHeadLimit = 8
)

func (c *Client) registryTagsForDigest(ctx context.Context, repoName, digest string) ([]string, error) {
	repo, err := name.NewRepository(repoName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error parsing repository")
	}

	allTags, err := remote.List(repo,
		remote.WithContext(ctx),
		remote.WithAuth(authn.Anonymous),
	)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing registry tags")
	}

	if len(allTags) > maxRegistryTags {
		allTags = allTags[len(allTags)-maxRegistryTags:]
	}

	var (
		mu      sync.Mutex
		matched = make([]string, 0)
		group   errgroup.Group
	)

	group.SetLimit(registryHeadLimit)

	for _, tag := range allTags {
		group.Go(func() error {
			if !hasTagDigest(ctx, repo.Tag(tag), digest) {
				return nil
			}

			mu.Lock()

			matched = append(matched, tag)
			mu.Unlock()

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, rerrors.Wrap(err, "error checking registry tags")
	}

	return matched, nil
}

func hasTagDigest(ctx context.Context, tag name.Tag, digest string) bool {
	desc, err := remote.Head(tag,
		remote.WithContext(ctx),
		remote.WithAuth(authn.Anonymous),
	)
	if err != nil {
		return false
	}

	return desc.Digest.String() == digest
}
