package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers/seeder"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	apiPrefix = "/api/v4"

	privateTokenHeader = "Private-Token"
	contentTypeHeader  = "Content-Type"
	formContentType    = "application/x-www-form-urlencoded"

	maxResponseBytes = 1 << 20

	runnerTypeProject = "project_type"
	runnerTypeGroup   = "group_type"

	defaultConcurrent = 1
)

var _ seeder.ConfigSeeder = (*Provider)(nil)

type idResponse struct {
	Id int64 `json:"id"`
}

type createdRunnerResponse struct {
	Id    int64  `json:"id"`
	Token string `json:"token"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type configDocument struct {
	Concurrent int32            `toml:"concurrent"`
	Runners    []runnerDocument `toml:"runners"`
}

type runnerDocument struct {
	Name     string         `toml:"name"`
	Url      string         `toml:"url"`
	Token    string         `toml:"token"`
	Executor string         `toml:"executor"`
	Docker   dockerDocument `toml:"docker"`
}

type dockerDocument struct {
	Image      string `toml:"image"`
	Privileged bool   `toml:"privileged"`
}

type apiResponse struct {
	status int
	body   []byte
}

func (p *Provider) CreateRunner(ctx context.Context, req seeder.CreateRunnerReq) (seeder.CreatedRunner, error) {
	base := apiBase(req.BaseUrl)

	runnerType, lookupPath, idField, err := scopeEndpoints(req.Scope)
	if err != nil {
		return seeder.CreatedRunner{}, rerrors.Wrap(err)
	}

	targetId, err := p.lookupTargetId(ctx, base+lookupPath+url.PathEscape(req.Target), req.PersonalAccessToken)
	if err != nil {
		return seeder.CreatedRunner{}, rerrors.Wrap(err, "error resolving gitlab runner target")
	}

	form := url.Values{}
	form.Set("runner_type", runnerType)
	form.Set(idField, strconv.FormatInt(targetId, 10))
	form.Set("description", req.Description)

	if len(req.Labels) > 0 {
		form.Set("tag_list", strings.Join(req.Labels, ","))
	} else {
		form.Set("run_untagged", "true")
	}

	resp, err := p.callApi(ctx, http.MethodPost, base+"/user/runners", req.PersonalAccessToken, form)
	if err != nil {
		return seeder.CreatedRunner{}, rerrors.Wrap(err, "error creating gitlab runner")
	}

	if resp.status != http.StatusCreated {
		return seeder.CreatedRunner{}, rerrors.Wrap(statusError(resp.status, false), "error creating gitlab runner")
	}

	var created createdRunnerResponse

	err = json.Unmarshal(resp.body, &created)
	if err != nil {
		return seeder.CreatedRunner{}, rerrors.Wrap(err, "error parsing gitlab create runner response")
	}

	if created.Token == "" {
		return seeder.CreatedRunner{}, rerrors.Wrap(errRunnerTokenMissing)
	}

	result := seeder.CreatedRunner{Id: created.Id, Token: created.Token}

	return result, nil
}

func (p *Provider) RenderConfig(req seeder.RenderConfigReq) ([]byte, error) {
	image := req.DockerImage
	if image == "" {
		image = defaultDockerImage
	}

	runner := runnerDocument{
		Name:     req.RunnerName,
		Url:      apiHost(req.BaseUrl),
		Token:    req.RunnerToken,
		Executor: registerExecutor,
		Docker:   dockerDocument{Image: image, Privileged: true},
	}

	document := configDocument{Concurrent: defaultConcurrent, Runners: []runnerDocument{runner}}

	config, err := toml.Marshal(document)
	if err != nil {
		return nil, rerrors.Wrap(err, "error rendering gitlab-runner config.toml")
	}

	defaults := gitlab_runner_config.DockerDefaults{
		Volumes:             []string{req.CacheVolumeName + ":" + cacheMountPath},
		PullPolicy:          []string{pullPolicyIfNotPresent},
		AllowedPullPolicies: []string{pullPolicyIfNotPresent, pullPolicyAlways},
	}

	config, err = gitlab_runner_config.SetDockerDefaults(config, defaults)
	if err != nil {
		return nil, rerrors.Wrap(err, "error applying docker defaults")
	}

	concurrent := req.Concurrent
	if concurrent <= 0 {
		concurrent = defaultConcurrent
	}

	config = gitlab_runner_config.SetConcurrent(config, concurrent)

	if len(req.JobEnvironment) > 0 {
		config, err = gitlab_runner_config.SetRunnerEnvironment(config, proxyenv.Keys(), req.JobEnvironment)
		if err != nil {
			return nil, rerrors.Wrap(err, "error applying job environment")
		}
	}

	return config, nil
}

func (p *Provider) DeleteRunnerByToken(ctx context.Context, baseUrl, runnerToken string) error {
	form := url.Values{}
	form.Set("token", runnerToken)

	resp, err := p.callApi(ctx, http.MethodDelete, apiBase(baseUrl)+"/runners", "", form)
	if err != nil {
		return rerrors.Wrap(err, "error deleting gitlab runner")
	}

	if resp.status == http.StatusNoContent || resp.status == http.StatusNotFound {
		return nil
	}

	return rerrors.Wrap(statusError(resp.status, false), "error deleting gitlab runner")
}

func (p *Provider) ResetRunnerToken(ctx context.Context, baseUrl, runnerToken string) (string, error) {
	form := url.Values{}
	form.Set("token", runnerToken)

	resp, err := p.callApi(ctx, http.MethodPost, apiBase(baseUrl)+"/runners/reset_authentication_token", "", form)
	if err != nil {
		return "", rerrors.Wrap(err, "error resetting gitlab runner token")
	}

	if resp.status != http.StatusCreated {
		return "", rerrors.Wrap(statusError(resp.status, false), "error resetting gitlab runner token")
	}

	var reset tokenResponse

	err = json.Unmarshal(resp.body, &reset)
	if err != nil {
		return "", rerrors.Wrap(err, "error parsing gitlab reset runner token response")
	}

	if reset.Token == "" {
		return "", rerrors.Wrap(errRunnerTokenMissing)
	}

	return reset.Token, nil
}

func (p *Provider) lookupTargetId(ctx context.Context, endpoint, accessToken string) (int64, error) {
	resp, err := p.callApi(ctx, http.MethodGet, endpoint, accessToken, nil)
	if err != nil {
		return 0, rerrors.Wrap(err, "error calling gitlab lookup endpoint")
	}

	if resp.status != http.StatusOK {
		return 0, rerrors.Wrap(statusError(resp.status, true))
	}

	var target idResponse

	err = json.Unmarshal(resp.body, &target)
	if err != nil {
		return 0, rerrors.Wrap(err, "error parsing gitlab lookup response")
	}

	if target.Id == 0 {
		return 0, rerrors.Wrap(errTargetIdMissing)
	}

	return target.Id, nil
}

// callApi sends the request with the personal access token when one is given
// (runner-token endpoints authenticate by the form's token instead).
func (p *Provider) callApi(
	ctx context.Context, method, endpoint, accessToken string, form url.Values,
) (apiResponse, error) {
	var body io.Reader

	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return apiResponse{}, rerrors.Wrap(err, "error building gitlab api request")
	}

	if accessToken != "" {
		httpReq.Header.Set(privateTokenHeader, accessToken)
	}

	if form != nil {
		httpReq.Header.Set(contentTypeHeader, formContentType)
	}

	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := p.httpClient().Do(httpReq)
	if err != nil {
		return apiResponse{}, rerrors.Wrap(err, "error calling gitlab api")
	}
	defer common.CloseWithLog(resp.Body.Close, "gitlab api response body")

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return apiResponse{}, rerrors.Wrap(err, "error reading gitlab api response")
	}

	result := apiResponse{status: resp.StatusCode, body: respBody}

	return result, nil
}

func (p *Provider) httpClient() *http.Client {
	if p.client != nil {
		return p.client
	}

	return defaultHttpClient
}

func scopeEndpoints(scope velez_api.RunnerScope) (runnerType, lookupPath, idField string, err error) {
	switch scope {
	case velez_api.RunnerScope_REPO:
		return runnerTypeProject, "/projects/", "project_id", nil
	case velez_api.RunnerScope_ORG:
		return runnerTypeGroup, "/groups/", "group_id", nil
	case velez_api.RunnerScope_RUNNER_SCOPE_UNSPECIFIED:
	default:
	}

	return "", "", "", rerrors.Wrap(errRunnerScopeUnsupported)
}

func statusError(status int, isLookup bool) error {
	detail := fmt.Sprintf("status: %d", status)

	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return rerrors.Wrap(user_errors.ErrGitlabAccessTokenRejected, detail)
	case isLookup && status == http.StatusNotFound:
		return rerrors.Wrap(user_errors.ErrGitlabTargetNotFound, detail)
	default:
		return rerrors.Wrap(errUnexpectedStatus, detail)
	}
}

func apiHost(baseUrl string) string {
	base := strings.TrimRight(baseUrl, "/")
	if base == "" {
		return defaultBaseUrl
	}

	return base
}

func apiBase(baseUrl string) string {
	return apiHost(baseUrl) + apiPrefix
}
