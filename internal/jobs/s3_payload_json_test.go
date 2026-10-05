package jobs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_CreateS3InstanceTaskPayload_JsonRoundTrip(t *testing.T) {
	t.Parallel()

	region := "ap-south"
	exposedPort := uint32(3900)
	adminPort := uint32(3903)
	webUiPort := uint32(3909)
	nodeId := "node-1"

	request := &velez_api.CreateS3Instance_Request{
		Name:        "files",
		Region:      &region,
		EnableWebUi: true,
	}

	payload := &velez_api.CreateS3InstanceTaskPayload{
		Request:          request,
		ExposedPort:      &exposedPort,
		AdminExposedPort: &adminPort,
		WebUiExposedPort: &webUiPort,
		NodeId:           &nodeId,
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	restored := &velez_api.CreateS3InstanceTaskPayload{}

	err = json.Unmarshal(raw, restored)
	require.NoError(t, err)

	require.Equal(t, "files", restored.GetRequest().GetName())
	require.Equal(t, region, restored.GetRequest().GetRegion())
	require.True(t, restored.GetRequest().GetEnableWebUi())
	require.Equal(t, exposedPort, restored.GetExposedPort())
	require.Equal(t, adminPort, restored.GetAdminExposedPort())
	require.Equal(t, webUiPort, restored.GetWebUiExposedPort())
	require.Equal(t, nodeId, restored.GetNodeId())
}

func Test_CreateRegistryInstanceTaskPayload_S3Storage_JsonRoundTrip(t *testing.T) {
	t.Parallel()

	bucketName := "images"
	bucketId := "bucket-1"
	accessKeyId := "GK123"

	request := &velez_api.CreateRegistryInstance_Request{
		Name: "reg",
		S3Storage: &velez_api.RegistryS3Storage{
			InstanceName: "files",
			BucketName:   &bucketName,
		},
	}

	payload := &velez_api.CreateRegistryInstanceTaskPayload{
		Request:       request,
		S3BucketId:    &bucketId,
		S3AccessKeyId: &accessKeyId,
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	restored := &velez_api.CreateRegistryInstanceTaskPayload{}

	err = json.Unmarshal(raw, restored)
	require.NoError(t, err)

	require.Equal(t, "files", restored.GetRequest().GetS3Storage().GetInstanceName())
	require.Equal(t, bucketName, restored.GetRequest().GetS3Storage().GetBucketName())
	require.Equal(t, bucketId, restored.GetS3BucketId())
	require.Equal(t, accessKeyId, restored.GetS3AccessKeyId())
}
