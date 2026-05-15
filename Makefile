.PHONY: setup
setup: codegen build-ui

codegen:
	@echo --- Generating contracts ---
	moti g
	@echo --- Generating sql queries ---
	sqlc generate

build-ui:
	@echo --- Building WebUI ---
	cd pkg/web/Velez-UI && bun && bun run build
	@echo --- Copying dist into Go embed path ---
	rm -rf internal/transport/ui/dist
	cp -r pkg/web/Velez-UI/dist internal/transport/ui/dist

lint:
	golangci-lint run ./...
