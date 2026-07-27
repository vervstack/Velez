package jobs

import (
	"context"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert/yaml"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	AssembleConfigAction = "assemble_config"

	// scratchContainerSuffix names the throwaway container used to read a
	// config out of an image, mirroring pipelines.configFetchingPostfix.
	scratchContainerSuffix = "_config_scanning"

	// stepFetchConfig names the fetch_config job shared with create_smerd.go and
	// upgrade_smerd.go's BuildJobs. stepDropContainer (same job name across
	// assemble_config.go and copy_to_volume.go) is declared once in
	// copy_to_volume.go and reused here.
	stepFetchConfig = "fetch_config"

	stepPrepareScratchImage = "prepare_image"

	// matreshka_api.ConfigTypePrefix's enum names. Kept as plain strings in the
	// TaskContext rather than the typed enum - see the conf_type field comment
	// in api/grpc/tasks.proto for why.
	confTypeVerv  = "verv"
	confTypePg    = "pg"
	confTypePlain = "plain"
)

// Accessor interfaces the assemble_config jobs need from their TaskContext.
// *velez_api.AssembleConfigTaskPayload satisfies all of them. containerIDAccessor
// is declared in create_smerd.go and reused here as-is.

type assembleConfigRequestAccessor interface {
	GetServiceName() string
	GetImageName() string
}

type imageMetaAccessor interface {
	GetImageLabels() map[string]string
	SetImageLabels(labels map[string]string)
	GetImageTags() []string
	SetImageTags(tags []string)
}

type configMetaAccessor interface {
	GetConfigName() string
	SetConfigName(configName string)
	GetConfType() string
	SetConfType(confType string)
	GetConfigFormat() velez_api.ConfigFormat
	SetConfigFormat(confFormat velez_api.ConfigFormat)
}

type configContentAccessor interface {
	GetContentRaw() []byte
	SetContentRaw(content []byte)
	SetContent(content []byte)
}

type assembleConfigHandler struct {
	nodeClients node_clients.NodeClients
}

func NewAssembleConfigHandler(nodeClients node_clients.NodeClients) TaskHandler {
	return &assembleConfigHandler{
		nodeClients: nodeClients,
	}
}

func (h *assembleConfigHandler) Action() string {
	return AssembleConfigAction
}

func (h *assembleConfigHandler) NewContext() TaskContext {
	return &velez_api.AssembleConfigTaskPayload{}
}

func (h *assembleConfigHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.AssembleConfigTaskPayload)
	if !ok {
		panic("assemble_config: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			Name: stepPrepareScratchImage,
			Job: &prepareScratchImageJob{
				docker: h.nodeClients.Docker(),
				req:    payload,
				ctx:    payload,
			},
		},
		{
			Name: stepCreatePgContainer,
			Job: &createScratchContainerJob{
				nodeClients: h.nodeClients,
				req:         payload,
				ctx:         payload,
			},
		},
		{
			Name: stepFetchConfig,
			Job: &fetchConfigJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				req:       payload,
				imageMeta: payload,
				container: payload,
				ctx:       payload,
				content:   payload,
			},
		},
		{
			Name: stepDropContainer,
			Job: &dropScratchContainerJob{
				docker: h.nodeClients.Docker(),
				ctx:    payload,
			},
		},
		{
			Name: "parse_config",
			Job: &parseConfigJob{
				req:     payload,
				content: payload,
			},
		},
	}
}

type prepareScratchImageJob struct {
	docker node_clients.Docker

	req assembleConfigRequestAccessor
	ctx imageMetaAccessor
}

func (j *prepareScratchImageJob) Do(ctx context.Context) error {
	imageInfo, err := j.docker.PullImage(ctx, j.req.GetImageName())
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	var imageLabels map[string]string

	if imageInfo.Config != nil {
		imageLabels = imageInfo.Config.Labels
	}

	j.ctx.SetImageLabels(imageLabels)
	j.ctx.SetImageTags(imageInfo.RepoTags)

	return nil
}

type createScratchContainerJob struct {
	nodeClients node_clients.NodeClients

	req assembleConfigRequestAccessor
	ctx containerIDAccessor
}

func (j *createScratchContainerJob) Do(ctx context.Context) error {
	name := j.req.GetServiceName() + scratchContainerSuffix

	cfg := &container.Config{
		Image:    j.req.GetImageName(),
		Hostname: name,
	}

	created, err := j.nodeClients.Docker().ContainerCreate(
		ctx, cfg, &container.HostConfig{}, &network.NetworkingConfig{}, &v1.Platform{}, name)
	if err != nil {
		return rerrors.Wrap(err, "error creating scratch container")
	}

	j.ctx.SetContainerId(created.ID)

	return nil
}

func (j *createScratchContainerJob) Rollback(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return nil
	}

	err := j.nodeClients.Docker().Remove(ctx, containerID)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrapf(err, "error removing scratch container '%s'", containerID)
	}

	return nil
}

type fetchConfigJob struct {
	dockerAPI client.APIClient

	req       assembleConfigRequestAccessor
	imageMeta imageMetaAccessor
	container containerIDAccessor
	ctx       configMetaAccessor
	content   configContentAccessor
}

// Do mirrors config_steps.fillMeta + getSpec + extractConfig, restricted to
// the shape AssembleConfig actually exercises: it never sets a Verv/Plain
// config override (unlike LaunchSmerd/UpgradeSmerd), so that branch of the
// original step is omitted here.
func (j *fetchConfigJob) Do(ctx context.Context) error {
	containerID := j.container.GetContainerId()
	if containerID == "" {
		return rerrors.New("empty container id")
	}

	confType, format, systemPath := classifyImage(j.imageMeta.GetImageLabels(), j.imageMeta.GetImageTags())

	j.ctx.SetConfigName(j.req.GetServiceName())
	j.ctx.SetConfType(confType)
	j.ctx.SetConfigFormat(format)

	if systemPath == "" {
		return nil
	}

	raw, err := dockerutils.ReadFromContainer(ctx, j.dockerAPI, containerID, systemPath)
	if err != nil {
		return rerrors.Wrap(err, "error getting config to mount")
	}

	j.content.SetContentRaw(raw)

	return nil
}

// classifyImage decides a config's type/format/system-path from an image's
// labels and tags, same rules as config_steps.fillMeta. Format is always
// "env" here since AssembleConfig never supplies a pre-set .yaml FilePath
// (the only way fillMeta would pick ConfigFormat_yaml).
func classifyImage(imageLabels map[string]string, imageTags []string) (
	confType string, format velez_api.ConfigFormat, systemPath string,
) {
	switch {
	case imageLabels[labels.MatreshkaConfigLabel] == vervConfigLabelEnabled:
		return confTypeVerv, velez_api.ConfigFormat_env, "/app/config/config.yaml"
	case isPostgresByImageTags(imageTags):
		return confTypePg, velez_api.ConfigFormat_env, ""
	default:
		return confTypePlain, velez_api.ConfigFormat_env, ""
	}
}

func isPostgresByImageTags(tags []string) bool {
	for _, tag := range tags {
		if strings.Contains(tag, "postgres") {
			return true
		}
	}

	return false
}

type dropScratchContainerJob struct {
	docker node_clients.Docker
	ctx    containerIDAccessor
}

func (j *dropScratchContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return nil
	}

	err := j.docker.Remove(ctx, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error dropping scratch container")
	}

	return nil
}

type parseConfigJob struct {
	req     configMetaAccessor
	content configContentAccessor
}

// Do mirrors do_assemble_config.go's getResult post-processing. Since jobs
// have no Runner.Result() equivalent, the parsed content is marshaled back
// to env-formatted bytes and stashed in the TaskContext's content field for
// the caller to read once the task is DONE.
func (j *parseConfigJob) Do(_ context.Context) error {
	raw := j.content.GetContentRaw()
	if len(raw) == 0 {
		return nil
	}

	var (
		parsed *evon.Node
		err    error
	)

	switch j.req.GetConfigFormat() {
	case velez_api.ConfigFormat_yaml:
		if j.req.GetConfType() == confTypeVerv {
			parsed, err = fromMatreshkaYamlToEvon(raw)
		} else {
			parsed, err = fromYamlToEvon(raw)
		}

		if err != nil {
			return rerrors.Wrap(err, "error parsing from yaml to evon")
		}
	case velez_api.ConfigFormat_env:
		err = evon.Unmarshal(raw, &parsed)
		if err != nil {
			return rerrors.Wrap(err, "error unmarshalling to evon")
		}
	}

	if parsed == nil {
		return nil
	}

	j.content.SetContent(evon.Marshal([]*evon.Node{parsed}))

	return nil
}

func fromMatreshkaYamlToEvon(content []byte) (*evon.Node, error) {
	c := matreshka.NewEmptyConfig()

	err := c.Unmarshal(content)
	if err != nil {
		return nil, rerrors.Wrap(err, "error unmarshalling to matreshka config")
	}

	env, err := evon.MarshalEnv(&c)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling to env")
	}

	return env, nil
}

func fromYamlToEvon(content []byte) (*evon.Node, error) {
	m := map[string]any{}

	err := yaml.Unmarshal(content, &m)
	if err != nil {
		return nil, rerrors.Wrap(err, "error unmarshalling from yaml to map")
	}

	env, err := evon.MarshalEnv(m)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling yaml to env")
	}

	return env, nil
}
