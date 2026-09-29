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

# tunnel-sc opens (if not already open) an SSH tunnel forwarding local port
# 12375 to sc's Docker socket, so serve-sc can point DOCKER_HOST at sc's
# Docker daemon instead of the local one.
tunnel-sc:
	@lsof -i :12375 >/dev/null 2>&1 || \
		(ssh -f -N -o ExitOnForwardFailure=yes -o ServerAliveInterval=30 -o ServerAliveCountMax=3 \
			-L 127.0.0.1:12375:/var/run/docker.sock sc && \
		 echo "--- Tunnel to sc's Docker opened on :12375 ---")

serve-sc: tunnel-sc
	@echo "--- Serving via Go backend (DOCKER_HOST -> sc via tunnel) + Vite dev server ---"
	@trap 'kill 0' EXIT INT TERM; \
	(cd pkg/web/Velez-UI && vite) & \
	(go build -o bin/velez-dev ./cmd/service && DOCKER_HOST=tcp://127.0.0.1:12375 ./bin/velez-dev --dev) & \
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