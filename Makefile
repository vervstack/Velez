.PHONY: setup
setup: codegen build-ui

codegen:
	@echo --- Generating contracts ---
	moti g
	@echo --- Generating sql queries ---
	sqlc generate

build-ui:
	@echo --- Building WebUI ---
	cd pkg/web/Velez-UI && bun i && bun run build
	@echo --- Copying dist into Go embed path ---
	rm -rf internal/transport/ui/dist
	cp -r pkg/web/Velez-UI/dist internal/transport/ui/dist

lint:
	golangci-lint run ./...

serve:
	@trap 'kill 0' EXIT; \
	go run ./cmd/service --dev & \
	cd pkg/web/Velez-UI && bun dev & \
	wait

client:
	cd pkg/web/Velez-UI && vite

# E2E suite. TestMain (tests/e2e/main_test.go) brings up a disposable
# Docker-in-Docker daemon, points DOCKER_HOST at it for the run, and tears
# it down afterwards. Needs a reachable bootstrap Docker daemon (the local
# socket, or VELEZ_E2E_DOCKER_HOST=tcp://host:port) able to start a
# privileged container.
#
# test-e2e is the always-green smoke deploy (Test_ContainerRuntime_Matrix
# only) that also runs in CI - every other suite carries a //go:build
# e2e_full tag so it doesn't compile into this binary. test-e2e-full is the
# same run with every suite included, for local use while working on e2e.
.PHONY: test-e2e
test-e2e:
	go test -count=1 -timeout 20m -parallel 4 ./tests/e2e/...

.PHONY: test-e2e-full
test-e2e-full:
	go test -tags e2e_full -count=1 -timeout 20m -parallel 4 ./tests/e2e/...