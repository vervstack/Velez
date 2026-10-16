//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

const s3SecretsRestartInstanceName = "e2e-s3-rst-sec"

func (s *S3InstanceSuite) Test_S3Instance_SecretsSurviveVelezRestart() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3SecretsRestartInstanceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, true)

	bucketReq := newCreateS3BucketRequest(instanceName, s3BucketName, s3OwnerService)

	_, err := env.Custom.S3ApiImpl.CreateS3Bucket(t.Context(), bucketReq)
	require.NoError(t, err)

	keyReq := newCreateS3KeyRequest(instanceName, s3ReadOnlyKeyName, s3BucketName)

	keyResp, err := env.Custom.S3ApiImpl.CreateS3Key(t.Context(), keyReq)
	require.NoError(t, err)
	require.NotEmpty(t, keyResp.GetSecretAccessKey())

	before := getS3InstanceCredentials(t, env, instanceName)
	require.NotEmpty(t, before.GetAdminToken())
	require.NotEmpty(t, before.GetWebUiPassword())

	forgetS3Secrets(t.Context(), t, env, instanceName, keyResp.GetAccessKeyId())

	after := getS3InstanceCredentials(t, env, instanceName)
	require.Equal(t, before.GetAdminToken(), after.GetAdminToken())
	require.Equal(t, before.GetWebUiPassword(), after.GetWebUiPassword())

	keyAfter := getS3KeyCredentials(t, env, instanceName, keyResp.GetAccessKeyId())
	require.Equal(t, keyResp.GetSecretAccessKey(), keyAfter.GetSecretAccessKey())

	rpcSecret, err := env.Custom.Services.Secrets().Get(t.Context(), domain.S3RpcSecretRef(instanceName))
	require.NoError(t, err)
	require.NotEmpty(t, rpcSecret)

	assertS3WebUiAnswers(t, after)

	dropS3Instance(t, env, instanceName)
}

func forgetS3Secrets(ctx context.Context, t *testing.T, env *TestEnvironment, instanceName, accessKeyId string) {
	t.Helper()

	refs := []domain.SecretRef{
		domain.S3AdminTokenSecretRef(instanceName),
		domain.S3RpcSecretRef(instanceName),
		domain.S3WebUiPasswordSecretRef(instanceName),
		domain.S3KeySecretRef(instanceName, accessKeyId),
	}

	for _, ref := range refs {
		err := env.Custom.Services.Secrets().Delete(ctx, ref)
		require.NoError(t, err)
	}
}
