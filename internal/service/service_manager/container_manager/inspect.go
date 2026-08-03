package container_manager

import (
	"context"
	"sort"
	"strings"
	"time"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// InspectSmerd resolves the ContainerRuntime serving environment and asks it
// to Inspect contID - same environment-scoped resolution/ownership semantics
// as ListSmerds (a container belonging to a different environment, or not
// found under either identifier form, is a real "not found" error here, not
// silently mixed up with a container from another environment). The
// returned InspectResponse's Name is already the virtual/logical name
// (labelBasedRuntime.Inspect rewrites it), so no suffix-stripping is needed
// here anymore.
func (c *ContainerManager) InspectSmerd(ctx context.Context, environment, contId string) (*velez_api.Smerd, error) {
	runtime, err := c.runtimes.Runtime(ctx, environment)
	if err != nil {
		return nil, errors.Wrap(err, "error resolving environment")
	}

	contInfo, found, err := runtime.Inspect(ctx, contId)
	if err != nil {
		return nil, errors.Wrap(err, "error inspecting container")
	}

	if !found {
		return nil, errors.New("container %q not found in environment %q", contId, environment)
	}

	bareName := strings.Replace(contInfo.Name, "/", "", 1)

	smerd := &velez_api.Smerd{
		Uuid:    contInfo.ID,
		Name:    bareName,
		Ports:   parser.ToPortsMapping(contInfo.HostConfig.PortBindings),
		Volumes: parser.ToVolume(contInfo.HostConfig.Mounts),
		Env:     parser.ToDockerEnv(contInfo.Config.Env),
		Labels:  contInfo.Config.Labels,
	}

	imageInfo, err := c.dockerAPI.ImageInspect(ctx, contInfo.Image)
	if err != nil {
		return nil, errors.Wrap(err, "error getting image info")
	}

	for _, imageName := range imageInfo.RepoTags[:1] {
		smerd.ImageName = imageName
	}

	if contInfo.State != nil {
		smerd.Status = velez_api.Smerd_Status(
			velez_api.Smerd_Status_value[contInfo.State.Status])
	}

	createdAt, err := time.Parse("2006-01-02T15:04:05Z", contInfo.Created)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing created at time")
	}

	smerd.CreatedAt = timestamppb.New(createdAt)

	for netName, net := range contInfo.NetworkSettings.Networks {
		nb := &velez_api.NetworkBind{
			NetworkName: netName,
		}

		if len(net.DNSNames) != 0 {
			nb.Aliases = make([]string, 0, len(net.DNSNames)-1)
		}

		for _, dName := range net.DNSNames {
			if !strings.HasPrefix(contInfo.ID, dName) {
				nb.Aliases = append(nb.Aliases, dName)
			}
		}

		smerd.Networks = append(smerd.Networks, nb)
	}

	sort.Slice(smerd.GetNetworks(), func(i, j int) bool {
		return smerd.GetNetworks()[i].GetNetworkName() < smerd.GetNetworks()[j].GetNetworkName()
	})

	return smerd, nil
}
