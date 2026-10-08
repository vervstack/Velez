package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_RedactSecrets_NestedMessageTokenIsRedactedAndOriginalUntouched(t *testing.T) {
	gitlab := &velez_api.GitlabConfig{AccessToken: "glpat-secret"}
	original := &velez_api.CreateRunner_Request{
		Name: "runner",
		ProviderConfig: &velez_api.CreateRunner_Request_Gitlab{
			Gitlab: gitlab,
		},
	}

	redacted := redactSecrets(original)

	redactedRequest, ok := redacted.(*velez_api.CreateRunner_Request)
	require.True(t, ok)
	require.Equal(t, "***", redactedRequest.GetGitlab().GetAccessToken())
	require.Equal(t, "runner", redactedRequest.GetName())
	require.Equal(t, "glpat-secret", original.GetGitlab().GetAccessToken())
	require.Equal(t, "glpat-secret", gitlab.GetAccessToken())
}

func Test_RedactSecrets_OptionalStringInNestedMessageIsRedacted(t *testing.T) {
	token := "glrt-secret"
	original := &velez_api.CreateRunnerTaskPayload{
		Request: &velez_api.CreateRunner_Request{
			ProviderConfig: &velez_api.CreateRunner_Request_Github{
				Github: &velez_api.GithubConfig{AccessToken: "ghp-secret"},
			},
		},
		RegistrationToken: &token,
	}

	redacted := redactSecrets(original)

	redactedPayload, ok := redacted.(*velez_api.CreateRunnerTaskPayload)
	require.True(t, ok)
	require.Equal(t, "***", redactedPayload.GetRegistrationToken())
	require.Equal(t, "***", redactedPayload.GetRequest().GetGithub().GetAccessToken())
	require.Equal(t, "glrt-secret", original.GetRegistrationToken())
	require.Equal(t, "ghp-secret", original.GetRequest().GetGithub().GetAccessToken())
}

func Test_RedactSecrets_EmptySecretStaysEmpty(t *testing.T) {
	original := &velez_api.GitlabConfig{}

	redacted := redactSecrets(original)

	redactedMessage, ok := redacted.(proto.Message)
	require.True(t, ok)
	require.True(t, proto.Equal(original, redactedMessage))
}

func Test_RedactSecrets_NonProtoPassesThrough(t *testing.T) {
	value := map[string]string{"access_token": "plain"}

	redacted := redactSecrets(value)

	require.Equal(t, value, redacted)
	require.Nil(t, redactSecrets(nil))
}
