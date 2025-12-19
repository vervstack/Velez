package pipelines

import (
	"go.vervstack.ru/Velez/internal/patterns"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
)

func (p *pipeliner) EnableStateFull() Runner[any] {
	//region Pipeline Context
	// TODO Generate
	password := "velez_pwd"
	containerName := "velez-state"
	launchContainer := patterns.Postgres(containerName, password)

	var containerId string
	//endregion

	return &runner[any]{
		Steps: []steps.Step{
			container_steps.Create(
				p.nodeClients, &launchContainer,
				&containerName, &containerId),
		},
		//	deploy postgres
		//	Connect this Velez to it
		//	Change settings in Velez's matreshka
	}
}
