RSCLI_VERSION=v0.0.30
rscli-version:
	@echo $(RSCLI_VERSION)

buildc:
	docker buildx build \
			--load \
			--platform linux/amd64,linux/arm64 \
			-t velez:local .