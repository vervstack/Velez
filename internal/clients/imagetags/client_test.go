package imagetags

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/distribution/reference"
	"github.com/stretchr/testify/require"
)

const (
	targetDigest = "sha256:aa"
)

func nginxRepository(t *testing.T) reference.Named {
	t.Helper()

	repository, err := reference.ParseNormalizedNamed("nginx")
	require.NoError(t, err)

	return repository
}

func newHub(t *testing.T, handler http.HandlerFunc) (*Client, *int32) {
	t.Helper()

	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	return NewWithHubUrl(srv.URL), &calls
}

func Test_TagsForDigest_MatchesTagDigest(t *testing.T) {
	client, _ := newHub(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/repositories/library/nginx/tags", r.URL.Path)

		_, _ = w.Write([]byte(`{"next": null, "results": [
			{"name": "latest", "digest": "sha256:aa"},
			{"name": "other", "digest": "sha256:bb"}]}`))
	})

	tags, err := client.TagsForDigest(t.Context(), nginxRepository(t), targetDigest)
	require.NoError(t, err)
	require.Equal(t, []string{"latest"}, tags)
}

func Test_TagsForDigest_MatchesImageDigest(t *testing.T) {
	client, _ := newHub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"next": null, "results": [
			{"name": "1.27.2", "digest": "sha256:cc", "images": [{"digest": "sha256:bb"}, {"digest": "sha256:aa"}]},
			{"name": "1.26", "digest": "sha256:dd", "images": [{"digest": "sha256:ee"}]}]}`))
	})

	tags, err := client.TagsForDigest(t.Context(), nginxRepository(t), targetDigest)
	require.NoError(t, err)
	require.Equal(t, []string{"1.27.2"}, tags)
}

func Test_TagsForDigest_FollowsNextPage(t *testing.T) {
	var srvUrl string

	client, calls := newHub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{"next": null, "results": [{"name": "1.27", "digest": "sha256:aa"}]}`))

			return
		}

		_, _ = w.Write([]byte(`{"next": "` + srvUrl + `/v2/repositories/library/nginx/tags?page=2", "results": [
			{"name": "latest", "digest": "sha256:aa"}]}`))
	})

	srvUrl = client.hubUrl

	tags, err := client.TagsForDigest(t.Context(), nginxRepository(t), targetDigest)
	require.NoError(t, err)
	require.Equal(t, []string{"latest", "1.27"}, tags)
	require.EqualValues(t, 2, atomic.LoadInt32(calls))
}

func Test_TagsForDigest_NoMatchReturnsEmpty(t *testing.T) {
	client, _ := newHub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"next": null, "results": [{"name": "latest", "digest": "sha256:zz"}]}`))
	})

	tags, err := client.TagsForDigest(t.Context(), nginxRepository(t), targetDigest)
	require.NoError(t, err)
	require.NotNil(t, tags)
	require.Empty(t, tags)
}

func Test_TagsForDigest_NonOkStatusReturnsError(t *testing.T) {
	client, _ := newHub(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := client.TagsForDigest(t.Context(), nginxRepository(t), targetDigest)
	require.ErrorIs(t, err, errHubUnexpectedStatus)
}

func Test_TagsForDigest_CacheHitMakesOneCall(t *testing.T) {
	client, calls := newHub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"next": null, "results": [{"name": "latest", "digest": "sha256:aa"}]}`))
	})

	first, err := client.TagsForDigest(t.Context(), nginxRepository(t), targetDigest)
	require.NoError(t, err)

	second, err := client.TagsForDigest(t.Context(), nginxRepository(t), targetDigest)
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.EqualValues(t, 1, atomic.LoadInt32(calls))
}
