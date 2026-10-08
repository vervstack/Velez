package gitlab

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers/seeder"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testAccessToken = "glpat-secret"
	testRunnerToken = "glrt-runner"
)

type recordedRequest struct {
	method      string
	escapedPath string
	header      http.Header
	form        url.Values
}

type gitlabStub struct {
	server   *httptest.Server
	requests []recordedRequest
}

func newGitlabStub(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *gitlabStub {
	t.Helper()

	stub := &gitlabStub{}

	stub.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		form, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		stub.requests = append(stub.requests, recordedRequest{
			method:      r.Method,
			escapedPath: r.URL.EscapedPath(),
			header:      r.Header.Clone(),
			form:        form,
		})

		handler(w, r)
	}))

	t.Cleanup(stub.server.Close)

	return stub
}

func respondJson(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)

	_, _ = w.Write([]byte(body))
}

func newCreateRunnerReq(
	baseUrl string, scope velez_api.RunnerScope, target string, labels []string,
) seeder.CreateRunnerReq {
	return seeder.CreateRunnerReq{
		BaseUrl:             baseUrl,
		PersonalAccessToken: testAccessToken,
		Scope:               scope,
		Target:              target,
		Description:         "velez runner",
		Labels:              labels,
	}
}

func Test_CreateRunner_RepoWithLabels(t *testing.T) {
	stub := newGitlabStub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			respondJson(w, http.StatusOK, `{"id": 42}`)

			return
		}

		respondJson(w, http.StatusCreated, `{"id": 7, "token": "`+testRunnerToken+`", "token_expires_at": null}`)
	})

	req := newCreateRunnerReq(stub.server.URL+"/", velez_api.RunnerScope_REPO, "group/sub/project", []string{"a", "b"})

	created, err := New().CreateRunner(t.Context(), req)

	require.NoError(t, err)
	require.Equal(t, seeder.CreatedRunner{Id: 7, Token: testRunnerToken}, created)
	require.Len(t, stub.requests, 2)

	lookup := stub.requests[0]
	require.Equal(t, http.MethodGet, lookup.method)
	require.Equal(t, "/api/v4/projects/group%2Fsub%2Fproject", lookup.escapedPath)
	require.Equal(t, testAccessToken, lookup.header.Get("Private-Token"))

	create := stub.requests[1]
	require.Equal(t, http.MethodPost, create.method)
	require.Equal(t, "/api/v4/user/runners", create.escapedPath)
	require.Equal(t, testAccessToken, create.header.Get("Private-Token"))
	require.Equal(t, "application/x-www-form-urlencoded", create.header.Get("Content-Type"))
	require.Equal(t, "project_type", create.form.Get("runner_type"))
	require.Equal(t, "42", create.form.Get("project_id"))
	require.Equal(t, "velez runner", create.form.Get("description"))
	require.Equal(t, "a,b", create.form.Get("tag_list"))
	require.False(t, create.form.Has("run_untagged"))
}

func Test_CreateRunner_RepoWithoutLabelsRunsUntagged(t *testing.T) {
	stub := newGitlabStub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			respondJson(w, http.StatusOK, `{"id": 42}`)

			return
		}

		respondJson(w, http.StatusCreated, `{"id": 7, "token": "`+testRunnerToken+`"}`)
	})

	_, err := New().CreateRunner(t.Context(), newCreateRunnerReq(stub.server.URL, velez_api.RunnerScope_REPO, "o/r", nil))

	require.NoError(t, err)

	create := stub.requests[1]
	require.Equal(t, "true", create.form.Get("run_untagged"))
	require.False(t, create.form.Has("tag_list"))
}

func Test_CreateRunner_OrgUsesGroupEndpoints(t *testing.T) {
	stub := newGitlabStub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			respondJson(w, http.StatusOK, `{"id": 5}`)

			return
		}

		respondJson(w, http.StatusCreated, `{"id": 9, "token": "`+testRunnerToken+`"}`)
	})

	req := newCreateRunnerReq(stub.server.URL, velez_api.RunnerScope_ORG, "parent/child", []string{"x"})

	created, err := New().CreateRunner(t.Context(), req)

	require.NoError(t, err)
	require.Equal(t, int64(9), created.Id)
	require.Equal(t, "/api/v4/groups/parent%2Fchild", stub.requests[0].escapedPath)
	require.Equal(t, "group_type", stub.requests[1].form.Get("runner_type"))
	require.Equal(t, "5", stub.requests[1].form.Get("group_id"))
	require.False(t, stub.requests[1].form.Has("project_id"))
}

func Test_CreateRunner_Failures(t *testing.T) {
	cases := []struct {
		name    string
		scope   velez_api.RunnerScope
		lookup  int
		create  int
		body    string
		wantErr error
	}{
		{"lookup unauthorized", velez_api.RunnerScope_REPO, http.StatusUnauthorized, 0, "",
			user_errors.ErrGitlabAccessTokenRejected},
		{"lookup not found", velez_api.RunnerScope_REPO, http.StatusNotFound, 0, "",
			user_errors.ErrGitlabTargetNotFound},
		{"lookup server error", velez_api.RunnerScope_REPO, http.StatusInternalServerError, 0, "", errUnexpectedStatus},
		{"create forbidden", velez_api.RunnerScope_REPO, http.StatusOK, http.StatusForbidden, "",
			user_errors.ErrGitlabAccessTokenRejected},
		{"create empty token", velez_api.RunnerScope_REPO, http.StatusOK, http.StatusCreated, `{"id": 1, "token": ""}`,
			errRunnerTokenMissing},
		{"unsupported scope", velez_api.RunnerScope_RUNNER_SCOPE_UNSPECIFIED, 0, 0, "", errRunnerScopeUnsupported},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := newGitlabStub(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					respondJson(w, tc.lookup, `{"id": 42}`)

					return
				}

				respondJson(w, tc.create, tc.body)
			})

			_, err := New().CreateRunner(t.Context(), newCreateRunnerReq(stub.server.URL, tc.scope, "o/r", nil))

			require.ErrorIs(t, err, tc.wantErr)
			require.NotContains(t, err.Error(), testAccessToken)
		})
	}
}

func Test_DeleteRunnerByToken_Statuses(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		wantErr error
	}{
		{"deleted", http.StatusNoContent, nil},
		{"already gone", http.StatusNotFound, nil},
		{"server error", http.StatusInternalServerError, errUnexpectedStatus},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := newGitlabStub(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			})

			err := New().DeleteRunnerByToken(t.Context(), stub.server.URL, testRunnerToken)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			require.Equal(t, http.MethodDelete, stub.requests[0].method)
			require.Equal(t, "/api/v4/runners", stub.requests[0].escapedPath)
			require.Equal(t, testRunnerToken, stub.requests[0].form.Get("token"))
		})
	}
}

func Test_ResetRunnerToken_ReturnsNewToken(t *testing.T) {
	stub := newGitlabStub(t, func(w http.ResponseWriter, _ *http.Request) {
		respondJson(w, http.StatusCreated, `{"token": "glrt-rotated", "token_expires_at": null}`)
	})

	token, err := New().ResetRunnerToken(t.Context(), stub.server.URL, testRunnerToken)

	require.NoError(t, err)
	require.Equal(t, "glrt-rotated", token)
	require.Equal(t, http.MethodPost, stub.requests[0].method)
	require.Equal(t, "/api/v4/runners/reset_authentication_token", stub.requests[0].escapedPath)
	require.Equal(t, testRunnerToken, stub.requests[0].form.Get("token"))
}

func Test_ResetRunnerToken_Failures(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{"server error", http.StatusInternalServerError, "", errUnexpectedStatus},
		{"empty token", http.StatusCreated, `{"token": ""}`, errRunnerTokenMissing},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := newGitlabStub(t, func(w http.ResponseWriter, _ *http.Request) {
				respondJson(w, tc.status, tc.body)
			})

			_, err := New().ResetRunnerToken(t.Context(), stub.server.URL, testRunnerToken)

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func newRenderConfigReq() seeder.RenderConfigReq {
	return seeder.RenderConfigReq{
		BaseUrl:         "https://gitlab.example.com/",
		RunnerToken:     testRunnerToken,
		RunnerName:      "my \"runner\"",
		DockerImage:     "golang:1.24",
		CacheVolumeName: "my-runner-cache",
		Concurrent:      4,
	}
}

func Test_RenderConfig_RoundTripsAndAppliesDefaults(t *testing.T) {
	out, err := New().RenderConfig(newRenderConfigReq())

	require.NoError(t, err)
	require.Equal(t, 1, gitlab_runner_config.CountRunners(out))

	concurrent, isFound := gitlab_runner_config.Concurrent(out)
	require.True(t, isFound)
	require.Equal(t, int32(4), concurrent)

	token, isFound := gitlab_runner_config.RunnerToken(out)
	require.True(t, isFound)
	require.Equal(t, testRunnerToken, token)

	settings, err := gitlab_runner_config.ReadSettings(out)
	require.NoError(t, err)
	require.Equal(t, []string{pullPolicyIfNotPresent}, settings.PullPolicy)
	require.Equal(t, []string{pullPolicyIfNotPresent, pullPolicyAlways}, settings.AllowedPullPolicies)

	text := string(out)
	require.Contains(t, text, `url = 'https://gitlab.example.com'`)
	require.Contains(t, text, `image = 'golang:1.24'`)
	require.Contains(t, text, `privileged = true`)
	require.Contains(t, text, `executor = 'docker'`)
	require.Contains(t, text, `"my-runner-cache:/cache"`)
	require.NotContains(t, text, "environment")
}

func Test_RenderConfig_DefaultsForEmptyFields(t *testing.T) {
	req := newRenderConfigReq()

	req.BaseUrl = ""
	req.DockerImage = ""
	req.Concurrent = 0

	out, err := New().RenderConfig(req)

	require.NoError(t, err)

	concurrent, _ := gitlab_runner_config.Concurrent(out)
	require.Equal(t, int32(1), concurrent)
	require.Contains(t, string(out), `url = 'https://gitlab.com'`)
	require.Contains(t, string(out), "image = '"+defaultDockerImage+"'")
}

func Test_RenderConfig_AppliesJobEnvironment(t *testing.T) {
	key := proxyenv.Keys()[0]

	req := newRenderConfigReq()

	req.JobEnvironment = map[string]string{key: "socks5://proxy:1080"}

	out, err := New().RenderConfig(req)

	require.NoError(t, err)
	require.Contains(t, string(out), key+"=socks5://proxy:1080")
	require.Equal(t, 1, gitlab_runner_config.CountRunners(out))
}
