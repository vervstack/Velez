// Package seeder declares the ConfigSeeder capability a runner provider can
// implement. Split out from the providers package so the provider
// implementations (gitlab) can assert it without an import cycle: providers
// imports every implementation.
package seeder

import (
	"context"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type ConfigSeeder interface {
	CreateRunner(ctx context.Context, req CreateRunnerReq) (CreatedRunner, error)
	RenderConfig(req RenderConfigReq) ([]byte, error)
	DeleteRunnerByToken(ctx context.Context, baseUrl, runnerToken string) error
	ResetRunnerToken(ctx context.Context, baseUrl, runnerToken string) (string, error)
}

type CreateRunnerReq struct {
	BaseUrl             string
	PersonalAccessToken string
	Scope               velez_api.RunnerScope
	Target              string
	Description         string
	Labels              []string
}

type CreatedRunner struct {
	Id    int64
	Token string
}

type RenderConfigReq struct {
	BaseUrl         string
	RunnerToken     string
	RunnerName      string
	DockerImage     string
	CacheVolumeName string
	Concurrent      int32
	JobEnvironment  map[string]string
}
