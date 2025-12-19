package pipelines

import (
	"go.vervstack.ru/Velez/internal/cluster/cluster_state"
	"go.vervstack.ru/Velez/internal/patterns/db_patterns/pg_pattern"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/smerd_steps"
)

func (p *pipeliner) EnableStatefullMode() Runner[any] {
	//region Pipeline Context
	// TODO Generate
	containerName := cluster_state.Name

	launchContainer := pg_pattern.Postgres(
		pg_pattern.WithInstanceName(containerName),
	)

	var containerId string
	//endregion

	return &runner[any]{
		Steps: []steps.Step{
			//	Deploy postgres
			container_steps.Create(
				p.nodeClients, &launchContainer.Pattern,
				&containerName, &containerId),
			smerd_steps.Start(p.nodeClients, &containerId),
			//	Connect this Velez to it
		},
		//	Change settings in Velez's matreshka
	}
}
