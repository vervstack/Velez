package pipelines

import (
	"path"

	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/smerd_steps"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (p *pipeliner) CopyToVolume(req domain.CopyToVolumeRequest) Runner[any] {
	return NewCopyToVolumeRunner(p.nodeClients, req)
}

func NewCopyToVolumeRunner(nodeClients node_clients.NodeClients, req domain.CopyToVolumeRequest) Runner[any] {
	// region Pipeline context

	baseContainer := domain.LaunchSmerd{
		CreateSmerd_Request: &velez_api.CreateSmerd_Request{
			Name:      req.VolumeName + "_loader",
			ImageName: "alpine",
			Command:   rtb.ToPtr("sleep infinity"),
			Settings: &velez_api.Container_Settings{
				Volumes: []*velez_api.Volume{},
			},
		},
	}

	var contId string

	var filesToMount []domain.FileMountPoint

	mountedFolders := map[string]struct{}{}
	for filePath, fileContent := range req.PathToFiles {
		filesToMount = append(filesToMount, domain.FileMountPoint{
			FilePath: rtb.ToPtr(filePath),
			Content:  fileContent,
		})

		fold := path.Dir(filePath)

		_, ok := mountedFolders[fold]
		if ok {
			continue
		}

		mountedFolders[fold] = struct{}{}

		baseContainer.Settings.Volumes = append(baseContainer.Settings.Volumes,
			&velez_api.Volume{
				VolumeName:    req.VolumeName,
				ContainerPath: fold,
			})
	}
	//endregion

	actualSteps := []steps.Step{
		smerd_steps.Create(nodeClients, &baseContainer, &contId),
		smerd_steps.Start(nodeClients, &contId),
	}

	for _, ftm := range filesToMount {
		actualSteps = append(actualSteps,
			smerd_steps.Exec(nodeClients, &contId, rtb.ToPtr("mkdir -p "+path.Dir(*ftm.FilePath))),
			container_steps.CopyToContainer(nodeClients, &contId, &ftm))
	}

	actualSteps = append(actualSteps, smerd_steps.DropContainerStep(nodeClients, &contId))
	return &runner[any]{
		Steps: actualSteps,
	}
}
