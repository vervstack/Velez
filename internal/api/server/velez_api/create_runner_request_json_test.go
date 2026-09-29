package velez_api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CreateRunnerRequestJson_GitlabConfigSurvivesRoundTrip(t *testing.T) {
	concurrent := int32(4)
	dockerImage := "alpine:3"

	original := &CreateRunner_Request{
		Name: "runner",
		ProviderConfig: &CreateRunner_Request_Gitlab{
			Gitlab: &GitlabConfig{
				AccessToken: "token",
				DockerImage: &dockerImage,
				Concurrent:  &concurrent,
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	restored := &CreateRunner_Request{}

	err = json.Unmarshal(data, restored)
	require.NoError(t, err)

	require.Equal(t, int32(4), restored.GetGitlab().GetConcurrent())
	require.Equal(t, dockerImage, restored.GetGitlab().GetDockerImage())
}
