// Package configresolver keeps one config per service: it lives in matreshka
// when the cluster has it, otherwise it travels on the CreateSmerd request and
// is read back from the running container.
//
// Only env config is stored in matreshka. The matreshka server cannot hold a
// raw file: a plain config cannot be created, SaveConfig accepts only kv and
// verv configs, and reading a kv config renders its nodes, not the bytes that
// were saved. Files therefore always ride on CreateSmerd_Request.Plain.
package configresolver

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type Resolver struct {
	configService service.ConfigurationService
	clusters      cluster_clients.ClusterClients
	runtimes      container_runtime.RuntimeResolver
}

func New(
	configService service.ConfigurationService,
	clusters cluster_clients.ClusterClients,
	runtimes container_runtime.RuntimeResolver,
) *Resolver {
	return &Resolver{
		configService: configService,
		clusters:      clusters,
		runtimes:      runtimes,
	}
}

// WriteEnv stores env for the service named request.Name. With matreshka it
// lands in the kv config of that name, which fetchSmerdConfigJob pulls at
// deploy through request.Verv; without it, it is merged into request.Env.
func (r *Resolver) WriteEnv(ctx context.Context, request *velez_api.CreateSmerd_Request, env map[string]string) error {
	if len(env) == 0 {
		return nil
	}

	configName := request.GetName()

	existing, state, err := r.loadMatreshkaEnv(ctx, configName)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if state == configStateDisabled {
		mergeEnvIntoRequest(request, env)

		return nil
	}

	err = validateEnvKeys(env)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = r.storeMatreshkaEnv(ctx, configName, state, existing, env)
	if err != nil {
		return rerrors.Wrap(err)
	}

	attachMatreshkaConfig(request, configName)

	return nil
}

// WriteFile always puts the file on the request, see the package comment.
func (r *Resolver) WriteFile(
	_ context.Context, request *velez_api.CreateSmerd_Request, path string, content []byte,
) error {
	for _, file := range request.GetPlain() {
		if file.GetPath() == path {
			file.Content = content

			return nil
		}
	}

	file := &velez_api.FileConfig{Path: path, Content: content}

	request.Plain = append(request.Plain, file)

	return nil
}

// ReadEnv reads from matreshka when it holds values for the service, else
// from the running container.
func (r *Resolver) ReadEnv(ctx context.Context, serviceName, environment string) (map[string]string, error) {
	meta := domain.ConfigMeta{Name: serviceName}

	root, err := r.configService.GetEnvFromApi(ctx, meta)
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading env from matreshka")
	}

	env := make(map[string]string)
	collectLeafValues(root, env)

	if len(env) > 0 {
		return env, nil
	}

	runtime, err := r.runtimes.Runtime(ctx, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	info, isFound, err := runtime.Inspect(ctx, serviceName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting service container")
	}

	if !isFound || info.Config == nil {
		return nil, rerrors.Wrap(user_errors.ErrNoSuchContainer)
	}

	for _, pair := range info.Config.Env {
		key, value, _ := strings.Cut(pair, "=")

		env[key] = value
	}

	return env, nil
}

func (r *Resolver) ReadFile(ctx context.Context, serviceName, environment, path string) ([]byte, error) {
	runtime, err := r.runtimes.Runtime(ctx, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	content, err := runtime.CopyFromContainer(ctx, serviceName, path)
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading file from service container")
	}

	return content, nil
}

func (r *Resolver) Delete(ctx context.Context, serviceName string) error {
	return r.deleteMatreshkaConfig(ctx, serviceName)
}

func mergeEnvIntoRequest(request *velez_api.CreateSmerd_Request, env map[string]string) {
	if request.Env == nil {
		request.Env = make(map[string]string, len(env))
	}

	maps.Copy(request.GetEnv(), env)
}

func collectLeafValues(node *evon.Node, out map[string]string) {
	if node == nil {
		return
	}

	if node.Name != "" && node.Value != nil {
		out[node.Name] = fmt.Sprint(node.Value)
	}

	for _, inner := range node.InnerNodes {
		collectLeafValues(inner, out)
	}
}
