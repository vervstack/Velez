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
	@echo --- Starting Go backend + Vite dev server ---
	@trap 'trap - EXIT INT TERM; kill 0' EXIT INT TERM; \
	(cd pkg/web/Velez-UI && vite; kill 0) & \
	(go build -o bin/velez-dev ./cmd/service && ./bin/velez-dev --dev; kill 0) & \
	wait

build-n-serve: build-ui
	@echo "--- Serving via Go (embedded UI) ---"
	go build -o bin/velez-dev ./cmd/service
	./bin/velez-dev --dev

bns: build-n-serve

# serve-sc talks to sc's Docker daemon directly over TCP; published container
# ports are dialed on the same host. dockerd on sc listens on 192.168.1.44:2375
# (daemon.json "hosts", LAN only - unauthenticated).
SC_DOCKER_HOST ?= tcp://192.168.1.44:2375

serve-sc:
	@echo "--- Serving via Go backend (DOCKER_HOST -> $(SC_DOCKER_HOST)) + Vite dev server ---"
	@trap 'trap - EXIT INT TERM; kill 0' EXIT INT TERM; \
	(cd pkg/web/Velez-UI && vite; kill 0) & \
	(go build -o bin/velez-dev ./cmd/service && DOCKER_HOST=$(SC_DOCKER_HOST) ./bin/velez-dev --dev; kill 0) & \
	wait

kill-port:
	@lsof -ti :53890 | xargs -r kill -9

# E2E suite. TestMain (tests/e2e/main_test.go) brings up a disposable
# Docker-in-Docker daemon, points DOCKER_HOST at it for the run, and tears
# it down afterwards. Needs a reachable bootstrap Docker daemon (the local
# socket, or VELEZ_E2E_DOCKER_HOST=tcp://host:port) able to start a
# privileged container.

# Builds the velez:e2e image (native cross-compiled binary on a cached alpine,
# no web pulls) that e2egroups loads into every group's daemon. prepare/run
# build it themselves; this target is for rebuilding on its own.
.PHONY: e2e-image
e2e-image:
	go run ./tests/e2egroups image

# Full parallel e2e: one DinD daemon and one test process per group in
# tests/e2e/groups.yaml (plus the implicit `rest` group), all groups at once,
# with Velez running as a container in each group's daemon. Needs a reachable
# bootstrap docker daemon like the other e2e targets.
#   e2e-prepare  warms every group's daemon (pull images, seed, networks);
#                optional, e2e does it for any group that is missing.
#   e2e          builds the image once, prepares missing daemons, runs all groups, at most
#                MAX_GROUPS at once (make e2e MAX_GROUPS=2; default from the docker daemon's memory).
#   e2e-clean    removes every group's daemon container and cache volume.
#   e2e-group    runs one group (GROUP=<name>), preparing its daemon first if missing.
.PHONY: e2e-prepare
e2e-prepare:
	go run ./tests/e2egroups prepare

.PHONY: e2e
e2e:
	go run ./tests/e2egroups run --prepare-missing $(if $(MAX_GROUPS),--max-groups $(MAX_GROUPS))

.PHONY: e2e-clean
e2e-clean:
	go run ./tests/e2egroups clean

# One group only: make e2e-group GROUP=matreshka (prepares it if missing).
.PHONY: e2e-group
e2e-group:
	@test -n "$(GROUP)" || { echo "GROUP is required, e.g. make e2e-group GROUP=matreshka"; exit 1; }
	go run ./tests/e2egroups run --group $(GROUP) --prepare-missing
