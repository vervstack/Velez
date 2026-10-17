//go:build e2e_full

package e2e

import (
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
	dockerClient := env.DockerClient()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, true)

	bucketReq := newCreateS3BucketRequest(instanceName, s3BucketName, s3OwnerService)

	_, err := env.S3Client().CreateS3Bucket(t.Context(), bucketReq)
	require.NoError(t, err)

	keyReq := newCreateS3KeyRequest(instanceName, s3ReadOnlyKeyName, s3BucketName)

	keyResp, err := env.S3Client().CreateS3Key(t.Context(), keyReq)
	require.NoError(t, err)
	require.NotEmpty(t, keyResp.GetSecretAccessKey())

	before := getS3InstanceCredentials(t, env, instanceName)
	require.NotEmpty(t, before.GetAdminToken())
	require.NotEmpty(t, before.GetWebUiPassword())

	env = restartedEnvironment(t, env)

	after := getS3InstanceCredentials(t, env, instanceName)
	require.Equal(t, before.GetAdminToken(), after.GetAdminToken())
	require.Equal(t, before.GetWebUiPassword(), after.GetWebUiPassword())

	keyAfter := getS3KeyCredentials(t, env, instanceName, keyResp.GetAccessKeyId())
	require.Equal(t, keyResp.GetSecretAccessKey(), keyAfter.GetSecretAccessKey())

	rpcSecret := env.SecretValue(t, domain.S3RpcSecretRef(instanceName))
	require.NotEmpty(t, rpcSecret)

	assertS3WebUiAnswers(t, after)

	dropS3Instance(t, env, instanceName)
}
