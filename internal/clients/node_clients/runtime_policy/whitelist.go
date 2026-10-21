package runtime_policy

import (
	"strings"
)

var elevatedAccessWhitelist = []string{
	"tailscale/tailscale",
	"docker",
	"gitlab/gitlab-runner",
	"ghcr.io/actions/actions-runner",
	"khairul169/garage-webui",
	"moby/buildkit",
	"tonistiigi/binfmt",
}

// BuiltinWhitelist returns a copy of the repositories always allowed to run with elevated access.
func BuiltinWhitelist() []string {
	return append([]string(nil), elevatedAccessWhitelist...)
}

// RepositoryOf returns the lowercase repository of an image reference without tag, digest or registry host - the
// form the user whitelist stores and matches on.
func RepositoryOf(image string) string {
	return stripRegistryHost(imageRepository(image))
}

func isImageWhitelisted(image string, userWhitelist []string) bool {
	repository := imageRepository(image)
	withoutRegistry := stripRegistryHost(repository)

	for _, allowed := range elevatedAccessWhitelist {
		if matchesRepository(repository, withoutRegistry, allowed) {
			return true
		}
	}

	for _, allowed := range userWhitelist {
		if matchesRepository(repository, withoutRegistry, allowed) {
			return true
		}
	}

	return false
}

func matchesRepository(repository, withoutRegistry, allowed string) bool {
	return repository == allowed ||
		withoutRegistry == allowed ||
		withoutRegistry == "library/"+allowed
}

func imageRepository(image string) string {
	repository, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(image)), "@")

	lastSlash := strings.LastIndex(repository, "/")
	lastColon := strings.LastIndex(repository, ":")

	if lastColon > lastSlash {
		repository = repository[:lastColon]
	}

	return repository
}

func stripRegistryHost(repository string) string {
	host, rest, ok := strings.Cut(repository, "/")
	if !ok {
		return repository
	}

	isHost := strings.ContainsAny(host, ".:") || host == "localhost"

	if !isHost {
		return repository
	}

	return rest
}
