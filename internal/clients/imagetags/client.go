package imagetags

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/distribution/reference"
	"go.redsock.ru/rerrors"
)

const (
	defaultHubUrl = "https://hub.docker.com"
	hubDomain     = "docker.io"
	cacheTtl      = 5 * time.Minute
	hubTimeout    = 10 * time.Second
)

type cacheEntry struct {
	tags      []string
	expiresAt time.Time
}

// Client finds the tags of an image repository that point at a given manifest digest.
type Client struct {
	hubUrl     string
	httpClient *http.Client

	mu    sync.Mutex
	cache map[string]cacheEntry
}

func New() *Client {
	return NewWithHubUrl(defaultHubUrl)
}

func NewWithHubUrl(hubUrl string) *Client {
	return &Client{
		hubUrl:     hubUrl,
		httpClient: &http.Client{Timeout: hubTimeout},
		cache:      make(map[string]cacheEntry),
	}
}

// TagsForDigest returns the tags of repository whose manifest digest equals digest ("sha256:...").
func (c *Client) TagsForDigest(ctx context.Context, repository reference.Named, digest string) ([]string, error) {
	key := repository.Name() + "@" + digest

	tags, isCached := c.getCached(key)
	if isCached {
		return tags, nil
	}

	var err error

	if reference.Domain(repository) == hubDomain {
		tags, err = c.hubTagsForDigest(ctx, reference.Path(repository), digest)
	} else {
		tags, err = c.registryTagsForDigest(ctx, repository.Name(), digest)
	}

	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving tags for digest")
	}

	c.putCached(key, tags)

	return append([]string{}, tags...), nil
}

func (c *Client) getCached(key string) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		delete(c.cache, key)

		return nil, false
	}

	return append([]string{}, entry.tags...), true
}

func (c *Client) putCached(key string, tags []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = cacheEntry{
		tags:      append([]string{}, tags...),
		expiresAt: time.Now().Add(cacheTtl),
	}
}
