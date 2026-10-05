package jobs

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/garage"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	stepEnsureS3Storage      = "ensure_s3_storage"
	stepRegisterS3Dependency = "register_s3_dependency"

	registryS3Protocol  = "s3"
	s3RegionConfigField = "s3_region"
)

type registryS3Accessor interface {
	GetS3BucketId() string
	GetS3AccessKeyId() string
}

func registryS3BucketName(request *velez_api.CreateRegistryInstance_Request) string {
	return toolbox.Coalesce(request.GetS3Storage().GetBucketName(), request.GetName())
}

// ensureS3StorageJob provisions the bucket and the read/write key the
// registry stores its blobs with. The bucket is reused when it already
// exists, so pointing a new registry at an existing bucket keeps its data.
type ensureS3StorageJob struct {
	runtimes container_runtime.RuntimeResolver
	secrets  secrets.Store
	payload  *velez_api.CreateRegistryInstanceTaskPayload
}

func (j *ensureS3StorageJob) Do(ctx context.Context) error {
	request := j.payload.GetRequest()
	environment := request.GetEnvironment()
	instance := request.GetS3Storage().GetInstanceName()

	err := j.ensureInstanceExists(ctx, environment, instance)
	if err != nil {
		return err
	}

	client, err := newGarageClient(ctx, j.runtimes, j.secrets, environment, instance)
	if err != nil {
		return err
	}

	bucketName := registryS3BucketName(request)

	bucketId, err := findOrCreateBucket(ctx, client, bucketName)
	if err != nil {
		return err
	}

	j.payload.S3BucketId = &bucketId

	return j.ensureKey(ctx, client, instance, bucketId, bucketName)
}

func (j *ensureS3StorageJob) ensureInstanceExists(ctx context.Context, environment, instance string) error {
	containerRuntime, err := j.runtimes.Runtime(ctx, environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	listReq := &velez_api.ListSmerds_Request{
		Environment: environment,
		Label:       map[string]string{labels.S3InstanceLabel: instance},
	}

	containers, err := containerRuntime.ListContainers(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing s3 instance containers")
	}

	if len(containers) == 0 {
		return rerrors.Wrap(user_errors.ErrS3InstanceNotFound)
	}

	return nil
}

func (j *ensureS3StorageJob) ensureKey(
	ctx context.Context,
	client *garage.Client,
	instance, bucketId, bucketName string,
) error {
	keyId := j.payload.GetS3AccessKeyId()

	if keyId == "" {
		keyName := j.payload.GetRequest().GetName() + ":" + bucketName

		key, err := client.CreateKey(ctx, keyName)
		if err != nil {
			return rerrors.Wrap(err, "error creating s3 key")
		}

		keyId = key.AccessKeyId

		err = j.secrets.Put(ctx, domain.S3KeySecretRef(instance, keyId), toolbox.FromPtr(key.SecretAccessKey))
		if err != nil {
			return rerrors.Wrap(err, "error putting s3 key secret")
		}

		j.payload.S3AccessKeyId = &keyId
	}

	permissions := garage.BucketKeyPerm{Read: true, Write: true}

	_, err := client.AllowBucketKey(ctx, bucketId, keyId, permissions)
	if err != nil {
		return rerrors.Wrap(err, "error allowing s3 key on bucket")
	}

	return nil
}

func findOrCreateBucket(ctx context.Context, client *garage.Client, alias string) (string, error) {
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		return "", rerrors.Wrap(err, "error listing s3 buckets")
	}

	for _, bucket := range buckets {
		if slices.Contains(bucket.GlobalAliases, alias) {
			return bucket.Id, nil
		}
	}

	created, err := client.CreateBucket(ctx, alias)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating s3 bucket")
	}

	return created.Id, nil
}

// readS3InstanceRegion reads the region back from the instance's own garage
// config (serviceName is the prefixed service name), so a registry signs its requests for the region the instance
// actually serves.
func readS3InstanceRegion(
	ctx context.Context,
	configResolver service.ServiceConfigResolver,
	serviceName, environment string,
) (string, error) {
	config, err := configResolver.ReadFile(ctx, serviceName, environment, domain.S3ConfigPath)
	if err != nil {
		return "", rerrors.Wrap(err, "error reading s3 instance config")
	}

	for line := range strings.SplitSeq(string(config), "\n") {
		field, value, isPair := strings.Cut(line, "=")
		if isPair && strings.TrimSpace(field) == s3RegionConfigField {
			return strings.Trim(strings.TrimSpace(value), `"`), nil
		}
	}

	return domain.S3DefaultRegion, nil
}

func appendDependsOnLabel(request *velez_api.CreateSmerd_Request, instance string) {
	if request.Labels == nil {
		request.Labels = make(map[string]string)
	}

	existing := request.GetLabels()[labels.DependsOnLabel]
	if existing == "" {
		request.Labels[labels.DependsOnLabel] = instance

		return
	}

	request.Labels[labels.DependsOnLabel] = existing + "," + instance
}

func (j *deployRegistryInstanceJob) overlayS3Storage(
	ctx context.Context,
	smerdRequest *velez_api.CreateSmerd_Request,
) error {
	request := j.req.GetRequest()
	instance := request.GetS3Storage().GetInstanceName()
	keyId := j.s3.GetS3AccessKeyId()

	region, err := readS3InstanceRegion(ctx, j.configResolver, domain.S3ServiceName(instance), request.GetEnvironment())
	if err != nil {
		return err
	}

	secretKey, err := j.secrets.Get(ctx, domain.S3KeySecretRef(instance, keyId))
	if err != nil {
		return rerrors.Wrap(err, "error reading s3 key secret")
	}

	plainEnv := map[string]string{
		domain.RegistryStorageEnv:            domain.RegistryStorageS3Value,
		domain.RegistryStorageS3RegionEnv:    region,
		domain.RegistryStorageS3EndpointEnv:  s3InternalUrl(domain.S3ServiceName(instance), domain.S3ApiContainerPort),
		domain.RegistryStorageS3BucketEnv:    registryS3BucketName(request),
		domain.RegistryStorageS3PathStyleEnv: strconv.FormatBool(true),
		domain.RegistryStorageS3AccessKeyEnv: keyId,
		// Docker clients can't reach the internal S3 endpoint a blob redirect would point them at.
		domain.RegistryStorageRedirectEnv: strconv.FormatBool(true),
	}

	err = j.configResolver.WriteEnv(ctx, smerdRequest, plainEnv)
	if err != nil {
		return rerrors.Wrap(err, "error writing registry s3 env")
	}

	secretEnv := map[string]string{domain.RegistryStorageS3SecretKeyEnv: secretKey}

	setRequestEnv(smerdRequest, secretEnv)
	attachS3Network(smerdRequest, instance)
	appendDependsOnLabel(smerdRequest, domain.S3ServiceName(instance))

	return nil
}

type registerRegistryS3DependencyJob struct {
	dependencies storage.ServiceDependenciesStorage
	source       string
	target       string
}

func (j *registerRegistryS3DependencyJob) Do(ctx context.Context) error {
	err := j.dependencies.UpsertDependency(ctx, j.source, j.target, registryS3Protocol)
	if err != nil {
		return rerrors.Wrap(err, "error registering registry s3 dependency")
	}

	return nil
}
