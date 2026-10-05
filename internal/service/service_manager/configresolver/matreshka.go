package configresolver

import (
	"context"
	"regexp"
	"sort"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type configState int

const (
	masterVersion = "master"
	// SaveConfig is the only call that can fill an empty kv config, and its
	// env parser splits on every '=', so it is seeded with a value that has
	// none and the real one is patched over it.
	seedValue = "seed"

	configStateDisabled configState = 0
	configStateMissing  configState = 1
	configStatePresent  configState = 2
)

// Matreshka upper-cases patched keys and rejects anything outside this set,
// so a key that differs from its upper-case form would be silently renamed.
var envKeyPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{2,}$`)

// loadMatreshkaEnv doubles as the matreshka-enabled check: a cluster without
// matreshka answers every call with ErrServiceIsDisabled.
func (r *Resolver) loadMatreshkaEnv(ctx context.Context, configName string) (map[string]string, configState, error) {
	req := &matreshka_api.GetConfigNode_Request{
		ConfigName: configName,
		Version:    masterVersion,
	}

	resp, err := r.clusters.Configurator().GetConfigNodes(ctx, req)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrServiceIsDisabled) {
			return nil, configStateDisabled, nil
		}

		if status.Code(err) == codes.NotFound {
			return nil, configStateMissing, nil
		}

		return nil, configStateDisabled, rerrors.Wrap(err, "error getting config nodes from matreshka")
	}

	existing := make(map[string]string)
	collectApiLeaves(resp.GetRoot(), existing)

	return existing, configStatePresent, nil
}

func collectApiLeaves(node *matreshka_api.Node, out map[string]string) {
	if node == nil {
		return
	}

	if node.GetName() != "" && node.Value != nil {
		out[node.GetName()] = node.GetValue()
	}

	for _, inner := range node.GetInnerNodes() {
		collectApiLeaves(inner, out)
	}
}

func validateEnvKeys(env map[string]string) error {
	for key := range env {
		if !envKeyPattern.MatchString(key) {
			return rerrors.Wrap(errInvalidEnvKey, key)
		}
	}

	return nil
}

func (r *Resolver) storeMatreshkaEnv(
	ctx context.Context,
	configName string,
	state configState,
	existing map[string]string,
	env map[string]string,
) error {
	client := r.clusters.Configurator()

	if state == configStateMissing {
		createReq := &matreshka_api.CreateConfig_Request{
			ConfigName: configName,
			ConfigType: matreshka_api.ConfigType_kv,
		}

		_, err := client.CreateConfig(ctx, createReq)
		if err != nil {
			return rerrors.Wrap(err, "error creating matreshka config")
		}
	}

	keys := sortedKeys(env)

	if len(existing) == 0 {
		saveReq := &matreshka_api.SaveConfig_Request{
			Format:     matreshka_api.Format_env,
			ConfigName: configName,
			Config:     []byte(keys[0] + "=" + seedValue + "\n"),
		}

		_, err := client.SaveConfig(ctx, saveReq)
		if err != nil {
			return rerrors.Wrap(err, "error seeding matreshka config")
		}
	}

	patchReq := &matreshka_api.PatchConfig_Request{
		ConfigName: configName,
		Patches:    buildPatches(keys, existing, env),
	}

	_, err := client.PatchConfig(ctx, patchReq)
	if err != nil {
		return rerrors.Wrap(err, "error patching matreshka config")
	}

	return nil
}

func buildPatches(keys []string, existing, env map[string]string) []*matreshka_api.Patch {
	patches := make([]*matreshka_api.Patch, 0, len(keys)+len(existing))

	for _, key := range keys {
		update := &matreshka_api.Patch_UpdateValue{UpdateValue: env[key]}

		patches = append(patches, &matreshka_api.Patch{FieldName: key, Patch: update})
	}

	for _, key := range sortedKeys(existing) {
		_, isKept := env[key]
		if isKept {
			continue
		}

		del := &matreshka_api.Patch_Delete{Delete: true}

		patches = append(patches, &matreshka_api.Patch{FieldName: key, Patch: del})
	}

	return patches
}

func (r *Resolver) deleteMatreshkaConfig(ctx context.Context, configName string) error {
	req := &matreshka_api.DeleteConfig_Request{ConfigName: configName}

	_, err := r.clusters.Configurator().DeleteConfig(ctx, req)
	if err == nil {
		return nil
	}

	if rerrors.Is(err, user_errors.ErrServiceIsDisabled) || status.Code(err) == codes.NotFound {
		return nil
	}

	return rerrors.Wrap(err, "error deleting matreshka config")
}

func attachMatreshkaConfig(request *velez_api.CreateSmerd_Request, configName string) {
	format := velez_api.ConfigFormat_env

	request.Verv = &velez_api.MatreshkaConfigSpec{
		ConfigName:   &configName,
		ConfigFormat: &format,
	}

	if request.Env == nil {
		request.Env = make(map[string]string)
	}

	request.IgnoreConfig = false
}

func sortedKeys(in map[string]string) []string {
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
