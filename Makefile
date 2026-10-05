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
	@trap 'kill 0' EXIT INT TERM; \
	(cd pkg/web/Velez-UI && vite) & \
	(go build -o bin/velez-dev ./cmd/service && ./bin/velez-dev --dev) & \
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
	@trap 'kill 0' EXIT INT TERM; \
	(cd pkg/web/Velez-UI && vite) & \
	(go build -o bin/velez-dev ./cmd/service && DOCKER_HOST=$(SC_DOCKER_HOST) ./bin/velez-dev --dev) & \
	wait

kill-port:
	@lsof -ti :53890 | xargs -r kill -9

# E2E suite. TestMain (tests/e2e/main_test.go) brings up a disposable
# Docker-in-Docker daemon, points DOCKER_HOST at it for the run, and tears
# it down afterwards. Needs a reachable bootstrap Docker daemon (the local
# socket, or VELEZ_E2E_DOCKER_HOST=tcp://host:port) able to start a
# privileged container.
#
# test-e2e is the always-green smoke deploy (Test_ContainerRuntime_Matrix
# only) - every other suite carries a //go:build e2e_full tag so it doesn't
# compile into this binary. test-e2e-full is the same run with every suite
# included; CI (branch-push.yaml's e2e-test job) runs this one.
.PHONY: test-e2e
test-e2e:
	go test -count=1 -timeout 20m -parallel 8 ./tests/e2e/...

.PHONY: test-e2e-full
test-e2e-full:
	go test -tags e2e_full -count=1 -timeout 20m -parallel 8 ./tests/e2e/...