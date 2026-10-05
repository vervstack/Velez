//go:build e2e_full

package e2e

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	s3LifecycleInstanceName  = "e2e-s3-lifecycle"
	s3ValidationInstanceName = "e2e-s3-validation"
	s3RegistryInstanceName   = "e2e-s3-registry-host"
	s3RegistryServiceName    = "e2e-s3-registry"

	s3MatreshkaInstanceName = "e2e-s3-mat"
	s3MatreshkaRegistryName = "e2e-s3-mat-registry"

	s3TaskTimeout     = 6 * time.Minute
	s3HttpTimeout     = 30 * time.Second
	s3CounterTimeout  = 90 * time.Second
	s3CounterPoll     = 2 * time.Second
	s3ObjectKey       = "hello.txt"
	s3ObjectBody      = "hello garage"
	s3ObjectMime      = "text/plain"
	s3DefaultRegion   = "garage"
	s3OwnerService    = "e2e-s3-owner-service"
	s3ReadOnlyKeyName = "e2e-read-only"
	s3BucketName      = "e2e-bucket"

	s3RegistryPushImage = "alpine"
	s3RegistryPushRepo  = "e2e/alpine"
	s3RegistryPushTag   = "s3"
)

// S3InstanceSuite exercises S3-as-a-Service (Garage) end to end in plain
// single-node/local_storage mode. CreateS3Instance only enqueues the
// create_s3_instance task; that task deploys Garage (and optionally the
// garage-webui sidecar) and waits for both containers itself, so a DONE task
// is already proof the containers are up.
type S3InstanceSuite struct {
	suite.Suite

	plane Plane
}

// planeName keeps a fixed test name unique per plane: every plane runs in
// parallel against the same Docker engine and the same shared matreshka.
func (s *S3InstanceSuite) planeName(base string) string {
	return base + "-" + string(s.plane.Mode)
}

func (s *S3InstanceSuite) Test_S3Instance_Lifecycle() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3LifecycleInstanceName)
	ownerService := s.planeName(s3OwnerService)

	ctx := t.Context()

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createReq := newCreateS3InstanceRequest(instanceName, true)

	createResp, err := env.Custom.S3ApiImpl.CreateS3Instance(ctx, createReq)
	require.NoError(t, err)

	waitForS3Task(t, env, createResp.GetEntityId(), createResp.GetAction())

	instance := findS3Instance(t, env, instanceName)
	require.NotNil(t, instance, "expected s3 instance %q in ListS3Instances", instanceName)
	require.NotZero(t, instance.GetS3Port())
	require.NotZero(t, instance.GetWebUiPort())
	require.Equal(t, s3DefaultRegion, instance.GetRegion())
	require.EqualValues(t, 1, instance.GetReplicationFactor())

	serviceReq := newGetServiceRequest(domain.S3ServiceName(instanceName))

	serviceResp, err := env.Custom.ServiceApiImpl.GetService(ctx, serviceReq)
	require.NoError(t, err, "the garage service must be addressable by its prefixed name")
	require.Equal(t, instanceName, serviceResp.GetAbout().GetOriginalName())

	instanceCreds := getS3InstanceCredentials(t, env, instanceName)
	require.NotEmpty(t, instanceCreds.GetAdminToken())
	require.NotEmpty(t, instanceCreds.GetS3Endpoint())
	require.Equal(t, "http://"+domain.S3ServiceName(instanceName)+":3900", instanceCreds.GetInternalS3Endpoint())
	require.Equal(t, s3DefaultRegion, instanceCreds.GetRegion())
	require.NotEmpty(t, instanceCreds.GetWebUiUrl())
	require.NotEmpty(t, instanceCreds.GetWebUiUsername())
	require.NotEmpty(t, instanceCreds.GetWebUiPassword())

	bucketReq := newCreateS3BucketRequest(instanceName, s3BucketName, ownerService)

	bucketResp, err := env.Custom.S3ApiImpl.CreateS3Bucket(ctx, bucketReq)
	require.NoError(t, err)
	require.Equal(t, s3BucketName, bucketResp.GetBucket().GetName())

	bucket := findS3Bucket(t, env, instanceName, s3BucketName)
	require.NotNil(t, bucket)
	require.Equal(t, ownerService, bucket.GetOwnerService())
	require.Len(t, bucket.GetAccess(), 1)

	ownerAccess := bucket.GetAccess()[0]
	require.True(t, ownerAccess.GetIsRead())
	require.True(t, ownerAccess.GetIsWrite())

	ownerCreds := getS3KeyCredentials(t, env, instanceName, ownerAccess.GetAccessKeyId())
	require.NotEmpty(t, ownerCreds.GetSecretAccessKey())

	ownerClient := newS3Client(t, ownerCreds.GetS3Endpoint(), ownerCreds.GetAccessKeyId(), ownerCreds.GetSecretAccessKey())

	putS3Object(t, ownerClient, s3BucketName)
	require.Equal(t, s3ObjectBody, getS3Object(t, ownerClient, s3BucketName))

	keyReq := newCreateS3KeyRequest(instanceName, s3ReadOnlyKeyName, s3BucketName)

	keyResp, err := env.Custom.S3ApiImpl.CreateS3Key(ctx, keyReq)
	require.NoError(t, err)
	require.NotEmpty(t, keyResp.GetAccessKeyId())
	require.NotEmpty(t, keyResp.GetSecretAccessKey())

	readOnlyCreds := getS3KeyCredentials(t, env, instanceName, keyResp.GetAccessKeyId())
	require.Equal(t, keyResp.GetSecretAccessKey(), readOnlyCreds.GetSecretAccessKey())

	readOnlyClient := newS3Client(t, readOnlyCreds.GetS3Endpoint(), readOnlyCreds.GetAccessKeyId(),
		readOnlyCreds.GetSecretAccessKey())

	require.Equal(t, s3ObjectBody, getS3Object(t, readOnlyClient, s3BucketName))

	_, putErr := readOnlyClient.PutObject(ctx, s3BucketName, "denied.txt", bytes.NewReader([]byte("x")), 1,
		minio.PutObjectOptions{})
	require.Error(t, putErr, "read-only key must not be able to PutObject")
	require.Equal(t, "AccessDenied", minio.ToErrorResponse(putErr).Code)

	accessReq := newSetS3BucketAccessRequest(instanceName, s3BucketName, keyResp.GetAccessKeyId(), true, true)

	_, err = env.Custom.S3ApiImpl.SetS3BucketAccess(ctx, accessReq)
	require.NoError(t, err)

	_, putErr = readOnlyClient.PutObject(ctx, s3BucketName, "granted.txt", bytes.NewReader([]byte("x")), 1,
		minio.PutObjectOptions{})
	require.NoError(t, putErr, "key granted write must be able to PutObject")

	assertS3WebUiAnswers(t, instanceCreds)

	_, err = env.Custom.S3ApiImpl.DeleteS3Bucket(ctx, newDeleteS3BucketRequest(instanceName, s3BucketName))
	require.Error(t, err)
	require.Equal(t, codes.FailedPrecondition, status.Code(err))

	deleteKeyReq := &velez_api.DeleteS3Key_Request{
		InstanceName: instanceName,
		AccessKeyId:  keyResp.GetAccessKeyId(),
	}

	_, err = env.Custom.S3ApiImpl.DeleteS3Key(ctx, deleteKeyReq)
	require.NoError(t, err)

	emptyS3Bucket(t, ownerClient, s3BucketName)

	_, err = env.Custom.S3ApiImpl.DeleteS3Bucket(ctx, newDeleteS3BucketRequest(instanceName, s3BucketName))
	require.NoError(t, err)

	require.Nil(t, findS3Bucket(t, env, instanceName, s3BucketName), "bucket still listed after delete")

	dropS3Instance(t, env, instanceName)

	require.Nil(t, findS3Instance(t, env, instanceName), "instance still listed after drop")

	credsReq := &velez_api.GetS3InstanceCredentials_Request{Name: instanceName}

	_, err = env.Custom.S3ApiImpl.GetS3InstanceCredentials(ctx, credsReq)
	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func (s *S3InstanceSuite) Test_S3Instance_Validation() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3ValidationInstanceName)

	ctx := t.Context()

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	t.Run("replication factor 2 is rejected", func(t *testing.T) {
		req := newCreateS3InstanceRequest(instanceName, false)
		req.ReplicationFactor = uint32Ptr(2)

		_, err := env.Custom.S3ApiImpl.CreateS3Instance(ctx, req)
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("empty name is rejected", func(t *testing.T) {
		req := newCreateS3InstanceRequest("", false)

		_, err := env.Custom.S3ApiImpl.CreateS3Instance(ctx, req)
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid names are rejected", func(t *testing.T) {
		for _, name := range []string{"Upper", "with_underscore", "-leading", "trailing-", "has space"} {
			req := newCreateS3InstanceRequest(name, false)

			_, err := env.Custom.S3ApiImpl.CreateS3Instance(ctx, req)
			require.Error(t, err, "name %q must be rejected", name)
			require.Equal(t, codes.InvalidArgument, status.Code(err), "name %q", name)
		}
	})

	t.Run("duplicate name is rejected", func(t *testing.T) {
		req := newCreateS3InstanceRequest(instanceName, false)

		resp, err := env.Custom.S3ApiImpl.CreateS3Instance(ctx, req)
		require.NoError(t, err)

		waitForS3Task(t, env, resp.GetEntityId(), resp.GetAction())

		_, err = env.Custom.S3ApiImpl.CreateS3Instance(ctx, req)
		require.Error(t, err)
		require.Equal(t, codes.AlreadyExists, status.Code(err))

		dropS3Instance(t, env, instanceName)
	})
}

func (s *S3InstanceSuite) Test_Registry_On_S3() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3RegistryInstanceName)
	registryName := s.planeName(s3RegistryServiceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	cleanup := func() {
		removeContainerRegistryInstance(dockerClient, registryName)
		removeS3Instance(dockerClient, instanceName)
	}

	cleanup()
	t.Cleanup(cleanup)

	runRegistryOnS3(t, env, dockerClient, instanceName, registryName, func(*testing.T) {})
}

// Test_Registry_On_S3_Matreshka runs the same registry-on-S3 flow with
// matreshka present: the registry's S3 env must be written into matreshka
// under the service's config name and reach the container from there.
func (s *S3InstanceSuite) Test_Registry_On_S3_Matreshka() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3MatreshkaInstanceName)
	registryName := s.planeName(s3MatreshkaRegistryName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t, WithMatreshka())
	dockerClient := env.Custom.NodeClients.Docker().Client()

	cleanup := func() {
		removeContainerRegistryInstance(dockerClient, registryName)
		removeS3Instance(dockerClient, instanceName)
	}

	cleanup()
	t.Cleanup(cleanup)

	configName := registryServiceName(registryName)

	t.Cleanup(func() {
		deleteReq := &matreshka_api.DeleteConfig_Request{ConfigName: configName}

		_, _ = env.Custom.ClusterClients.Configurator().DeleteConfig(context.Background(), deleteReq)
	})

	verify := func(t *testing.T) {
		t.Helper()

		assertRegistryS3EnvInMatreshka(t, env, instanceName, registryName, configName)
		assertRegistryS3EnvInContainer(t, dockerClient, instanceName, registryName, configName)
	}

	runRegistryOnS3(t, env, dockerClient, instanceName, registryName, verify)
}

func Test_S3Instance(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &S3InstanceSuite{plane: plane}
	})
}

func runRegistryOnS3(
	t *testing.T,
	env *TestEnvironment,
	dockerClient client.APIClient,
	s3InstanceName, registryName string,
	verifyDeployed func(t *testing.T),
) {
	t.Helper()

	ctx := t.Context()

	s3Resp, err := env.Custom.S3ApiImpl.CreateS3Instance(ctx, newCreateS3InstanceRequest(s3InstanceName, false))
	require.NoError(t, err)

	waitForS3Task(t, env, s3Resp.GetEntityId(), s3Resp.GetAction())

	registryReq := newCreateRegistryOnS3Request(registryName, s3InstanceName)

	_, err = env.Custom.ContainerRegistryApiImpl.CreateRegistryInstance(ctx, registryReq)
	require.NoError(t, err)

	waitForRegistryInstanceDeploy(t, env, registryName, false)

	registryServiceName := registryServiceName(registryName)

	instance := findRegistryInstance(t, env, registryServiceName)
	require.NotNil(t, instance)
	require.Equal(t, s3InstanceName, instance.GetS3Instance())
	require.Equal(t, registryName, instance.GetS3Bucket())

	verifyDeployed(t)

	credsReq := &velez_api.GetRegistryInstanceCredentials_Request{Name: registryServiceName}

	credsResp, err := env.Custom.ContainerRegistryApiImpl.GetRegistryInstanceCredentials(ctx, credsReq)
	require.NoError(t, err)

	pushImageToRegistry(t, dockerClient, instance.GetPort(), credsResp.GetUsername(), credsResp.GetPassword())

	require.Eventually(t, func() bool {
		bucket := findS3Bucket(t, env, s3InstanceName, registryName)

		return bucket != nil && bucket.GetObjectCount() > 0
	}, s3CounterTimeout, s3CounterPoll, "registry blobs never showed up in bucket %q", registryName)

	_, err = env.Custom.S3ApiImpl.DropS3Instance(ctx, &velez_api.DropS3Instance_Request{Name: s3InstanceName})
	require.Error(t, err)
	require.Equal(t, codes.FailedPrecondition, status.Code(err))

	dropReq := &velez_api.DropRegistryInstance_Request{Name: registryServiceName}

	_, err = env.Custom.ContainerRegistryApiImpl.DropRegistryInstance(ctx, dropReq)
	require.NoError(t, err)

	dropS3Instance(t, env, s3InstanceName)
}

func wantRegistryS3Env(s3InstanceName, bucketName string) map[string]string {
	return map[string]string{
		domain.RegistryStorageEnv:            domain.RegistryStorageS3Value,
		domain.RegistryStorageS3BucketEnv:    bucketName,
		domain.RegistryStorageS3RegionEnv:    s3DefaultRegion,
		domain.RegistryStorageS3PathStyleEnv: "true",
		domain.RegistryStorageS3EndpointEnv:  "http://" + domain.S3ServiceName(s3InstanceName) + ":3900",
	}
}

func assertRegistryS3EnvInMatreshka(t *testing.T, env *TestEnvironment, s3InstanceName, registryName, configName string) {
	t.Helper()

	req := &matreshka_api.GetConfigNode_Request{ConfigName: configName, Version: "master"}

	resp, err := env.Custom.ClusterClients.Configurator().GetConfigNodes(t.Context(), req)
	require.NoError(t, err, "registry S3 env must be stored in matreshka under %q", configName)

	stored := make(map[string]string)
	collectMatreshkaLeaves(resp.GetRoot(), stored)

	for key, want := range wantRegistryS3Env(s3InstanceName, registryName) {
		require.Equal(t, want, stored[key], "matreshka value of %s", key)
	}

	require.NotEmpty(t, stored[domain.RegistryStorageS3AccessKeyEnv])
	require.NotContains(t, stored, domain.RegistryStorageS3SecretKeyEnv, "the S3 secret key must never be stored in matreshka")
}

func assertRegistryS3EnvInContainer(
	t *testing.T, dockerClient client.APIClient, s3InstanceName, registryName, containerName string,
) {
	t.Helper()

	info, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)

	for key, want := range wantRegistryS3Env(s3InstanceName, registryName) {
		require.Contains(t, info.Config.Env, key+"="+want)
	}

	var hasSecret bool

	for _, pair := range info.Config.Env {
		hasSecret = hasSecret || strings.HasPrefix(pair, domain.RegistryStorageS3SecretKeyEnv+"=")
	}

	require.True(t, hasSecret, "the S3 secret key must still reach the container")
}

func collectMatreshkaLeaves(node *matreshka_api.Node, out map[string]string) {
	if node == nil {
		return
	}

	if node.GetName() != "" && node.Value != nil {
		out[node.GetName()] = node.GetValue()
	}

	for _, inner := range node.GetInnerNodes() {
		collectMatreshkaLeaves(inner, out)
	}
}

func pushImageToRegistry(t *testing.T, dockerClient client.APIClient, port uint32, username, password string) {
	t.Helper()

	ctx := t.Context()

	registryHost := "localhost:" + strconv.Itoa(int(port))
	targetRef := registryHost + "/" + s3RegistryPushRepo + ":" + s3RegistryPushTag

	err := dockerClient.ImageTag(ctx, s3RegistryPushImage, targetRef)
	require.NoError(t, err)

	pushOpts := newPushOptions(t, username, password)

	reader, err := dockerClient.ImagePush(ctx, targetRef, pushOpts)
	require.NoError(t, err)

	defer func() { _ = reader.Close() }()

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotContains(t, string(output), `"errorDetail"`, "docker push failed: %s", string(output))
}

func waitForS3Task(t *testing.T, env *TestEnvironment, entityId, action string) {
	t.Helper()

	require.NotEmpty(t, entityId)
	require.Equal(t, jobs.CreateS3InstanceAction, action)

	ctx, cancel := context.WithTimeout(t.Context(), s3TaskTimeout)
	defer cancel()

	var task tasks_queries.VelezTask

	for observed := range env.Custom.JobsEngine.Watch(ctx, entityId, action) {
		task = observed
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status,
		"%s task for %q error: %s", action, entityId, task.Error.String)
}

func findS3Instance(t *testing.T, env *TestEnvironment, name string) *velez_api.S3Instance {
	t.Helper()

	listResp, err := env.Custom.S3ApiImpl.ListS3Instances(t.Context(), &velez_api.ListS3Instances_Request{})
	require.NoError(t, err)

	for _, instance := range listResp.GetInstances() {
		if instance.GetName() == name {
			return instance
		}
	}

	return nil
}

func findS3Bucket(t *testing.T, env *TestEnvironment, instanceName, bucketName string) *velez_api.S3Bucket {
	t.Helper()

	listReq := &velez_api.ListS3Buckets_Request{InstanceName: instanceName}

	listResp, err := env.Custom.S3ApiImpl.ListS3Buckets(t.Context(), listReq)
	require.NoError(t, err)

	for _, bucket := range listResp.GetBuckets() {
		if bucket.GetName() == bucketName {
			return bucket
		}
	}

	return nil
}

func getS3InstanceCredentials(
	t *testing.T, env *TestEnvironment, name string,
) *velez_api.GetS3InstanceCredentials_Response {
	t.Helper()

	req := &velez_api.GetS3InstanceCredentials_Request{Name: name}

	resp, err := env.Custom.S3ApiImpl.GetS3InstanceCredentials(t.Context(), req)
	require.NoError(t, err)

	return resp
}

func getS3KeyCredentials(
	t *testing.T, env *TestEnvironment, instanceName, accessKeyId string,
) *velez_api.GetS3KeyCredentials_Response {
	t.Helper()

	req := &velez_api.GetS3KeyCredentials_Request{InstanceName: instanceName, AccessKeyId: accessKeyId}

	resp, err := env.Custom.S3ApiImpl.GetS3KeyCredentials(t.Context(), req)
	require.NoError(t, err)

	return resp
}

func dropS3Instance(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	_, err := env.Custom.S3ApiImpl.DropS3Instance(t.Context(), &velez_api.DropS3Instance_Request{Name: name})
	require.NoError(t, err)
}

func newCreateS3InstanceRequest(name string, enableWebUi bool) *velez_api.CreateS3Instance_Request {
	return &velez_api.CreateS3Instance_Request{Name: name, EnableWebUi: enableWebUi}
}

func newCreateS3BucketRequest(instanceName, bucketName, ownerService string) *velez_api.CreateS3Bucket_Request {
	return &velez_api.CreateS3Bucket_Request{
		InstanceName: instanceName,
		BucketName:   bucketName,
		OwnerService: &ownerService,
	}
}

func newDeleteS3BucketRequest(instanceName, bucketName string) *velez_api.DeleteS3Bucket_Request {
	return &velez_api.DeleteS3Bucket_Request{InstanceName: instanceName, BucketName: bucketName}
}

func newCreateS3KeyRequest(instanceName, keyName, bucketName string) *velez_api.CreateS3Key_Request {
	access := &velez_api.S3BucketAccess{BucketName: bucketName, IsRead: true}

	return &velez_api.CreateS3Key_Request{
		InstanceName: instanceName,
		KeyName:      keyName,
		Access:       []*velez_api.S3BucketAccess{access},
	}
}

func newSetS3BucketAccessRequest(
	instanceName, bucketName, accessKeyId string, isRead, isWrite bool,
) *velez_api.SetS3BucketAccess_Request {
	access := &velez_api.S3BucketAccess{
		BucketName:  bucketName,
		AccessKeyId: accessKeyId,
		IsRead:      isRead,
		IsWrite:     isWrite,
	}

	return &velez_api.SetS3BucketAccess_Request{InstanceName: instanceName, Access: access}
}

func newCreateRegistryOnS3Request(name, s3InstanceName string) *velez_api.CreateRegistryInstance_Request {
	storage := &velez_api.RegistryS3Storage{InstanceName: s3InstanceName}

	return &velez_api.CreateRegistryInstance_Request{Name: name, S3Storage: storage}
}

func newPushOptions(t *testing.T, username, password string) image.PushOptions {
	t.Helper()

	authConfig := registry.AuthConfig{Username: username, Password: password}

	encoded, err := registry.EncodeAuthConfig(authConfig)
	require.NoError(t, err)

	return image.PushOptions{RegistryAuth: encoded}
}

func uint32Ptr(value uint32) *uint32 {
	return &value
}

// newS3Client dials a Garage S3 endpoint as reported by Velez
// ("http://localhost:<port inside the DinD>") through the bootstrap-host
// address the DinD publishes that port on.
func newS3Client(t *testing.T, endpoint, accessKeyId, secretAccessKey string) *minio.Client {
	t.Helper()

	hostAddr := endpointDindAddr(t, endpoint)

	s3Client, err := minio.New(hostAddr, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretAccessKey, ""),
		Region: s3DefaultRegion,
		Secure: false,
	})
	require.NoError(t, err)

	return s3Client
}

func endpointDindAddr(t *testing.T, endpoint string) string {
	t.Helper()

	parsed, err := url.Parse(endpoint)
	require.NoError(t, err)

	port, err := strconv.Atoi(parsed.Port())
	require.NoError(t, err, "endpoint %q has no port", endpoint)

	addr, ok := sharedDind.Addr(port)
	require.True(t, ok, "dind did not publish port %d of endpoint %q", port, endpoint)

	return addr
}

func putS3Object(t *testing.T, s3Client *minio.Client, bucketName string) {
	t.Helper()

	body := []byte(s3ObjectBody)
	opts := minio.PutObjectOptions{ContentType: s3ObjectMime}

	_, err := s3Client.PutObject(t.Context(), bucketName, s3ObjectKey, bytes.NewReader(body), int64(len(body)), opts)
	require.NoError(t, err)
}

func getS3Object(t *testing.T, s3Client *minio.Client, bucketName string) string {
	t.Helper()

	object, err := s3Client.GetObject(t.Context(), bucketName, s3ObjectKey, minio.GetObjectOptions{})
	require.NoError(t, err)

	defer func() { _ = object.Close() }()

	body, err := io.ReadAll(object)
	require.NoError(t, err)

	return string(body)
}

func emptyS3Bucket(t *testing.T, s3Client *minio.Client, bucketName string) {
	t.Helper()

	ctx := t.Context()

	listOpts := minio.ListObjectsOptions{Recursive: true}

	for object := range s3Client.ListObjects(ctx, bucketName, listOpts) {
		require.NoError(t, object.Err)

		err := s3Client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
		require.NoError(t, err)
	}
}

func assertS3WebUiAnswers(t *testing.T, creds *velez_api.GetS3InstanceCredentials_Response) {
	t.Helper()

	hostAddr := endpointDindAddr(t, creds.GetWebUiUrl())

	ctx, cancel := context.WithTimeout(t.Context(), s3HttpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+hostAddr+"/", nil)
	require.NoError(t, err)

	req.SetBasicAuth(creds.GetWebUiUsername(), creds.GetWebUiPassword())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// removeS3Instance force-removes the Garage container, its web UI sidecar and
// the instance's network and volumes, ignoring "no such ..." - mirrors
// removeContainerRegistryInstance.
func removeS3Instance(dockerClient client.APIClient, name string) {
	ctx := context.Background()
	removeOpts := container.RemoveOptions{Force: true}

	_ = dockerClient.ContainerRemove(ctx, domain.S3ServiceName(name), removeOpts)
	_ = dockerClient.ContainerRemove(ctx, domain.S3WebUiServiceName(name), removeOpts)
	_ = dockerClient.VolumeRemove(ctx, domain.S3MetaVolumeName(name), true)
	_ = dockerClient.VolumeRemove(ctx, domain.S3DataVolumeName(name), true)
	_ = dockerClient.NetworkRemove(ctx, domain.S3NetworkName(name))
}
