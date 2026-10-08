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

func isImageWhitelisted(image string) bool {
	repository := imageRepository(image)
	withoutRegistry := stripRegistryHost(repository)

	for _, allowed := range elevatedAccessWhitelist {
		if repository == allowed || withoutRegistry == allowed {
			return true
		}

		if withoutRegistry == "library/"+allowed {
			return true
		}
	}

	return false
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
