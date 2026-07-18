package jobs

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"sort"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
)

const CopyToVolumeAction = "copy_to_volume"

// loaderContainerSuffix mirrors do_copy_to_volume.go's "<volume>_loader"
// container name.
const loaderContainerSuffix = "_loader"

// Accessor interfaces the copy_to_volume jobs need from their TaskContext.
// *velez_api.CopyToVolumeTaskPayload satisfies all of them. containerIdAccessor
// is declared in create_smerd.go and reused here as-is.

type copyToVolumeRequestAccessor interface {
	GetVolumeName() string
	GetPathToFiles() map[string][]byte
}

// startAPI is the narrow slice of client.APIClient startLoaderContainerJob
// needs (container start, plus stop for its Rollback). node_clients.Docker
// has no Start method, so this job must reach the raw Docker engine client -
// same as create_smerd.go's startContainerJob. Keeping the dependency as this
// small interface (rather than client.APIClient directly) makes it
// hand-fakeable in unit tests without minimock: client.APIClient's method
// set is a superset of startAPI's, so the real client still satisfies it.
type startAPI interface {
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
}

// copyAPI is the narrow slice of client.APIClient copyFileJob needs to
// write a file into the loader container. See startAPI's doc comment for why
// this stays a small local interface rather than depending on client.APIClient.
type copyAPI interface {
	CopyToContainer(ctx context.Context, containerID, dstPath string, content io.Reader, options container.CopyToContainerOptions) error
}

type copyToVolumeHandler struct {
	nodeClients node_clients.NodeClients
}

func NewCopyToVolumeHandler(nodeClients node_clients.NodeClients) TaskHandler {
	return &copyToVolumeHandler{
		nodeClients: nodeClients,
	}
}

func (h *copyToVolumeHandler) Action() string {
	return CopyToVolumeAction
}

func (h *copyToVolumeHandler) NewContext() TaskContext {
	return &velez_api.CopyToVolumeTaskPayload{}
}

// BuildJobs runs fresh on every task run and every resume, and DONE jobs are
// skipped by (task_id, job_name). That means the mapping from copy_file_<n>
// to an actual file MUST be identical across calls, or a resumed task will
// checkpoint the wrong file as done. Go map iteration order is randomized,
// so sortedFilePaths sorts payload.GetPathToFiles()'s keys before the loop
// below assigns indices - see TestCopyToVolumeHandler_BuildJobs_Deterministic.
func (h *copyToVolumeHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload := taskCtx.(*velez_api.CopyToVolumeTaskPayload)

	pathToFiles := payload.GetPathToFiles()
	sortedPaths := sortedFilePaths(pathToFiles)
	folders := mountedFolders(payload.GetVolumeName(), sortedPaths)

	namedJobs := make([]NamedJob, 0, len(sortedPaths)+3)

	createJob := &createLoaderContainerJob{
		nodeClients: h.nodeClients,
		req:         payload,
		folders:     folders,
		ctx:         payload,
	}
	namedJobs = append(namedJobs, NamedJob{Name: "create_container", Job: createJob})

	startJob := &startLoaderContainerJob{
		dockerAPI: h.nodeClients.Docker().Client(),
		ctx:       payload,
	}
	namedJobs = append(namedJobs, NamedJob{Name: "start_container", Job: startJob})

	for i, filePath := range sortedPaths {
		copyJob := &copyFileJob{
			docker:   h.nodeClients.Docker(),
			copyAPI:  h.nodeClients.Docker().Client(),
			ctx:      payload,
			filePath: filePath,
			content:  pathToFiles[filePath],
		}

		name := fmt.Sprintf("copy_file_%d", i)
		namedJobs = append(namedJobs, NamedJob{Name: name, Job: copyJob})
	}

	dropJob := &dropLoaderContainerJob{
		docker: h.nodeClients.Docker(),
		ctx:    payload,
	}
	namedJobs = append(namedJobs, NamedJob{Name: "drop_container", Job: dropJob})

	return namedJobs
}

// sortedFilePaths returns m's keys sorted lexicographically. Used instead of
// ranging m directly wherever job identity/order must stay stable across
// BuildJobs calls (Go randomizes map iteration order on every call).
func sortedFilePaths(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// mountedFolders derives one *velez_api.Volume per distinct parent directory
// across sortedPaths, preserving first-seen order - mirrors
// do_copy_to_volume.go's mountedFolders map-based dedup. Requires its input
// pre-sorted (via sortedFilePaths) so the resulting volume order is stable.
func mountedFolders(volumeName string, sortedPaths []string) []*velez_api.Volume {
	seen := map[string]struct{}{}
	volumes := make([]*velez_api.Volume, 0, len(sortedPaths))

	for _, p := range sortedPaths {
		folder := path.Dir(p)

		_, ok := seen[folder]
		if ok {
			continue
		}
		seen[folder] = struct{}{}

		volume := &velez_api.Volume{
			VolumeName:    volumeName,
			ContainerPath: folder,
		}
		volumes = append(volumes, volume)
	}

	return volumes
}

// createLoaderContainerJob mirrors smerd_steps.Create restricted to what the
// loader container needs: an alpine image running "sleep infinity" with the
// volume's distinct parent directories mounted. Unlike smerd_steps.Create /
// create_smerd.go's createContainerJob, it never calls ContainerInspect after
// create - the created container's ID is used as-is, following
// assemble_config.go's createScratchContainerJob precedent, so this job
// depends only on the narrow node_clients.Docker interface and stays fully
// hand-fakeable.
type createLoaderContainerJob struct {
	nodeClients node_clients.NodeClients

	req     copyToVolumeRequestAccessor
	folders []*velez_api.Volume
	ctx     containerIdAccessor
}

func (j *createLoaderContainerJob) Do(ctx context.Context) error {
	name := j.req.GetVolumeName() + loaderContainerSuffix

	cfg := &container.Config{
		Image:    "alpine",
		Hostname: name,
		Cmd:      []string{"sleep", "infinity"},
	}

	settings := &velez_api.Container_Settings{Volumes: j.folders}
	mounts := parser.FromVolume(settings)

	hostCfg := &container.HostConfig{
		Mounts: mounts,
	}
	if len(j.folders) != 0 {
		hostCfg.VolumeDriver = j.folders[0].VolumeName
	}

	netCfg := &network.NetworkingConfig{}
	platform := &v1.Platform{}

	dockerClient := j.nodeClients.Docker()

	created, err := dockerClient.ContainerCreate(ctx, cfg, hostCfg, netCfg, platform, name)
	if err != nil {
		return rerrors.Wrap(err, "error creating loader container")
	}

	j.ctx.SetContainerId(created.ID)

	return nil
}

func (j *createLoaderContainerJob) Rollback(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	err := j.nodeClients.Docker().Remove(ctx, containerId)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrapf(err, "error removing loader container '%s'", containerId)
	}

	return nil
}

type startLoaderContainerJob struct {
	dockerAPI startAPI

	ctx containerIdAccessor
}

func (j *startLoaderContainerJob) Do(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return rerrors.New("no container id provided")
	}

	startOpts := container.StartOptions{}

	err := j.dockerAPI.ContainerStart(ctx, containerId, startOpts)
	if err != nil {
		return rerrors.Wrap(err, "error starting loader container")
	}

	return nil
}

func (j *startLoaderContainerJob) Rollback(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	stopOpts := container.StopOptions{}

	err := j.dockerAPI.ContainerStop(ctx, containerId, stopOpts)
	if err != nil {
		return rerrors.Wrapf(err, "error stopping loader container '%s'", containerId)
	}

	return nil
}

// copyFileJob combines the pipeline's per-file "mkdir -p <dir>" Exec step and
// its CopyToContainer step into a single checkpointed unit (see
// docs/jobs_migrations/questions.md #2) - mkdir -p is idempotent, so
// re-running it on task resume costs nothing. It has no Rollback: the
// pipeline never undid individual file writes either, only the container
// itself (via drop_container).
type copyFileJob struct {
	docker  node_clients.Docker
	copyAPI copyAPI

	ctx containerIdAccessor

	filePath string
	content  []byte
}

func (j *copyFileJob) Do(ctx context.Context) error {
	// Preserve container_steps.CopyToContainer's existing behavior of
	// silently skipping mount points whose content is nil - inherited
	// unchanged by this migration (see questions.md #3).
	if j.content == nil {
		return nil
	}

	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return rerrors.New("no container id provided")
	}

	dir := path.Dir(j.filePath)

	execOpts := container.ExecOptions{
		Cmd:    []string{"mkdir", "-p", dir},
		Detach: false,
	}

	// smerd_steps.Exec ignores the command's exit code (ops result is
	// discarded), so a failing mkdir surfaces only as a transport-level
	// Docker error here too - inherited unchanged (see questions.md #5).
	_, err := j.docker.Exec(ctx, containerId, execOpts)
	if err != nil {
		return rerrors.Wrap(err, "error creating directory in container")
	}

	err = writeFileToContainer(ctx, j.copyAPI, containerId, j.filePath, j.content)
	if err != nil {
		return rerrors.Wrap(err, "error copying file to container")
	}

	return nil
}

// writeFileToContainer mirrors dockerutils.WriteToContainer's tar-based
// write, but is parameterized on the narrow copyAPI interface rather than
// the full client.APIClient dockerutils.WriteToContainer requires. Go's
// interface assignability rules mean a fake satisfying only copyAPI can
// never be passed where dockerutils.WriteToContainer expects a full
// client.APIClient, so its logic is duplicated here rather than reused,
// trading a small amount of duplication for unit-testability without
// minimock.
func writeFileToContainer(ctx context.Context, dockerAPI copyAPI, contId string, systemPath string, content []byte) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name:    path.Base(systemPath),
		Mode:    0644,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}

	err := tw.WriteHeader(hdr)
	if err != nil {
		return rerrors.Wrap(err, "error writing tar header")
	}

	_, err = tw.Write(content)
	if err != nil {
		return rerrors.Wrap(err, "error writing content")
	}

	err = tw.Close()
	if err != nil {
		return rerrors.Wrap(err, "error closing tar writer")
	}

	copyOpts := container.CopyToContainerOptions{}

	err = dockerAPI.CopyToContainer(ctx, contId, path.Dir(systemPath), buf, copyOpts)
	if err != nil {
		return rerrors.Wrap(err, "error writing to container")
	}

	return nil
}

// dropLoaderContainerJob mirrors dropScratchContainerJob's shape.
type dropLoaderContainerJob struct {
	docker node_clients.Docker
	ctx    containerIdAccessor
}

func (j *dropLoaderContainerJob) Do(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	err := j.docker.Remove(ctx, containerId)
	if err != nil {
		return rerrors.Wrap(err, "error dropping loader container")
	}

	return nil
}
