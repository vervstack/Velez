package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

type Docker struct {
	directApi       client.APIClient
	bakedLabels     []string
	containerSuffix string
}

func NewClient(bakedLabels []string, containerSuffix string) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting docker client")
	}

	closer.Add(cli.Close)

	return &Docker{
		directApi:       cli,
		bakedLabels:     bakedLabels,
		containerSuffix: containerSuffix,
	}, nil
}

func (d *Docker) PullImage(ctx context.Context, imageName string) (image.InspectResponse, error) {
	_, err := dockerutils.PullImage(ctx, d.directApi, imageName, false)
	if err != nil {
		return image.InspectResponse{}, rerrors.Wrap(err, "error pulling image")
	}

	img, err := d.directApi.ImageInspect(ctx, imageName)
	if err != nil {
		return image.InspectResponse{}, rerrors.Wrap(err, "error inspecting image")
	}

	return img, nil
}

func (d *Docker) Remove(ctx context.Context, contUUID string) error {
	roReq := container.RemoveOptions{
		Force: true,
	}

	err := d.directApi.ContainerRemove(ctx, contUUID, roReq)
	if err != nil {
		if strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}
		return rerrors.Wrap(err, "error removing container")
	}

	return nil
}

func (d *Docker) Stop(ctx context.Context, nameOrId string) error {
	stopOpts := container.StopOptions{}
	err := d.directApi.ContainerStop(ctx, nameOrId, stopOpts)
	if err != nil {
		if strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}
		return rerrors.Wrap(err, "error stopping container")
	}

	return nil
}

func (d *Docker) Restart(ctx context.Context, nameOrId string) error {
	restartOpts := container.StopOptions{}
	err := d.directApi.ContainerRestart(ctx, nameOrId, restartOpts)
	if err != nil {
		if strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}
		return rerrors.Wrap(err, "error restarting container")
	}

	return nil
}

func (d *Docker) IsContainerRunning(ctx context.Context, nameOrId string) (bool, bool, error) {
	resp, err := d.directApi.ContainerInspect(ctx, nameOrId)
	if err != nil {
		if strings.Contains(err.Error(), NoSuchContainerError) {
			return false, false, nil
		}
		return false, false, rerrors.Wrap(err, "error inspecting container")
	}
	return resp.State.Running, true, nil
}

func (d *Docker) ListOccupiedPorts(ctx context.Context) ([]uint32, error) {
	containerList, err := d.directApi.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	usedPorts := make([]uint32, 0)
	for _, c := range containerList {
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				usedPorts = append(usedPorts, uint32(p.PublicPort))
			}
		}
	}

	return usedPorts, nil
}

func (d *Docker) ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error) {
	if req.Label == nil {
		req.Label = map[string]string{}
	}

	req.Label[labels.SuffixLabel] = d.containerSuffix

	list, err := dockerutils.ListContainers(ctx, d.directApi, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	return list, nil
}

func (d *Docker) Exec(ctx context.Context, containerId string, execCfg container.ExecOptions) ([]byte, error) {
	execResp, err := d.directApi.ContainerExecCreate(ctx, containerId, execCfg)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	if !execCfg.AttachStdout && !execCfg.AttachStderr {
		return nil, nil
	}

	// Attach to execution
	attachResp, err := d.directApi.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err)
	}
	defer attachResp.Close()

	dataOut := bytes.NewBuffer(nil)

	_, err = io.Copy(dataOut, attachResp.Reader)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return asciiSymbolsOnly(dataOut.Bytes()), nil
}

func (d *Docker) Client() client.APIClient {
	return d.directApi
}

func asciiSymbolsOnly(in []byte) []byte {
	cleanBuff := bytes.NewBuffer(nil)
	for _, b := range in {
		if b >= 32 && b <= 127 || b == '\n' {
			cleanBuff.WriteByte(b)
		}
	}

	return cleanBuff.Bytes()
}

func (d *Docker) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	config.Labels[labels.CreatedWithVelezLabel] = "true"
	config.Labels[labels.SuffixLabel] = d.containerSuffix

	for _, label := range d.bakedLabels {
		sepIdx := strings.Index(label, "=")
		name := label
		var val string
		if sepIdx != -1 {
			val = label[sepIdx+1:]
			name = label[:sepIdx]
		}

		config.Labels[name] = val
	}

	createResponse, err := d.directApi.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
	if err != nil {
		if errdefs.IsConflict(err) {
			return container.CreateResponse{}, handleConflictMessage(err)
		}
		return container.CreateResponse{}, rerrors.Wrap(err)
	}

	return createResponse, nil
}

func (d *Docker) Stats(ctx context.Context, nameOrId string) (domain.ContainerStats, error) {
	resp, err := d.directApi.ContainerStats(ctx, nameOrId, false)
	if err != nil {
		return domain.ContainerStats{}, rerrors.Wrap(err, "error getting container stats")
	}
	defer resp.Body.Close()

	var stats struct {
		CPUStats struct {
			CPUUsage struct {
				TotalUsage  uint64   `json:"total_usage"`
				PercpuUsage []uint64 `json:"percpu_usage"`
			} `json:"cpu_usage"`
			SystemUsage uint64 `json:"system_cpu_usage"`
			OnlineCPUs  int64  `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
	}

	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return domain.ContainerStats{}, rerrors.Wrap(err, "error decoding stats")
	}

	cpuPercent := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if cpuDelta > 0 && systemDelta > 0 {
		numCPU := stats.CPUStats.OnlineCPUs
		if numCPU == 0 {
			numCPU = int64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}
		cpuPercent = (cpuDelta / systemDelta) * float64(numCPU) * 100.0
	}

	memUsageMi := stats.MemoryStats.Usage / (1024 * 1024)
	memLimitMi := stats.MemoryStats.Limit / (1024 * 1024)

	inspectResp, err := d.directApi.ContainerInspect(ctx, nameOrId)
	if err != nil {
		return domain.ContainerStats{}, rerrors.Wrap(err, "error inspecting container")
	}

	var startedAt time.Time
	startedAtStr := inspectResp.State.StartedAt
	if startedAtStr != "" {
		t, err := time.Parse(time.RFC3339, startedAtStr)
		if err == nil {
			startedAt = t
		}
	}

	return domain.ContainerStats{
		CPUPercent: cpuPercent,
		MemUsageMi: uint64(memUsageMi),
		MemLimitMi: uint64(memLimitMi),
		StartedAt:  startedAt,
	}, nil
}
