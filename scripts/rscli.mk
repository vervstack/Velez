RSCLI_VERSION=v0.0.30
rscli-version:
	@echo $(RSCLI_VERSION)

buildc:
	docker build -t velez:local --no-cache .