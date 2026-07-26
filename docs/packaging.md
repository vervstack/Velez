# Packaging Velez for apt, Homebrew, and Chocolatey

Status: **planning document only**. Nothing described here is implemented yet — there is no
`.goreleaser.yml`, no `debian/` folder, no systemd unit, no Homebrew tap repo, and no Chocolatey
package in this repository today. This doc exists to inform decisions and give the repo owner a
concrete, actionable plan; it deliberately stops short of adding any packaging config so those
config files (`.goreleaser.yml`, a `velez.service` unit, a `nuspec`, etc.) can be reviewed and
added as an explicit follow-up.

## What the repo actually does today (grounding for everything below)

- **Distribution channel today is Docker only.** `.github/workflows/release.yaml` builds/pushes a
  Docker image on every push to `master` (via `RedSockActions/release_image`), tags a release
  (`RedSockActions/release_tag`), publishes the TypeScript client to npm, and renders
  `scripts/prod_init.sh.pattern` → `scripts/init_node.sh` (substituting the new version) and
  uploads it to MinIO/S3. `README.md`'s only install instructions are
  `wget -qO- scripts.vervstack.ru/init_node.sh | bash` and a raw `docker run`. There is no apt/deb,
  Homebrew, or Chocolatey step anywhere in CI.
- **`cmd/service/main.go` is a plain foreground process.** It calls `app.New()` then `a.Start()`
  and does nothing OS-service-specific. `go.mod` has no service-manager library — no
  `kardianos/service`, no `golang.org/x/sys/windows/svc` import (`golang.org/x/sys` is present only
  as an indirect dependency, pulled in by the Docker SDK, not used for service registration). So on
  every platform, "run Velez as a service" today means "have the OS supervise a plain binary,"
  not "the binary manages its own service lifecycle."
- **No systemd unit file exists anywhere in the repo** (`find . -iname "*.service"` returns
  nothing under application code).
- **The binary is already cheap to cross-compile.** The `Dockerfile` builds with
  `GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build ...` — a fully static, cross-compilable
  binary with no cgo dependencies. That's exactly the shape `goreleaser` wants, and it's a big part
  of why the recommendation below converges on one tool.
- **Config loading has a real gotcha for native packaging.** `internal/config/load.go` defaults to
  reading `./config/config.yaml` **relative to the process's working directory** (overridable with
  `-config <path>`; `-dev` additionally loads `./config/dev.yaml`). In the Docker image this works
  because the Dockerfile copies `config/` next to the binary in `/app` and the container's
  `WORKDIR` is `/app`. A native package has no such guarantee — a systemd/launchd/Windows-service
  unit that doesn't set the working directory (or pass `-config` explicitly) will silently load
  nothing or the wrong file. **Every native packaging path below must explicitly pass
  `-config /etc/velez/config.yaml`** (exact path depends on the decision in the callout section);
  don't rely on the default.
- **Velez needs privileged host access regardless of packaging format.** Per `README.md`, it needs
  `/var/run/docker.sock`, `/dev/disk`, and `/run`. It also uses `github.com/jaypipes/ghw` for
  hardware discovery, which typically needs elevated read access to `/sys`/`/proc`/SMART data. For
  a *native* (non-containerized) install this means the service user needs at minimum membership in
  the `docker` group (for socket access) and, in practice, may need root for full hardware
  discovery — this is a real security/scope trade-off, not a packaging detail, and is called out
  again below as something only the repo owner should decide.
- **Versioning is currently split across two sources of truth**: a hand-maintained
  `app_info.version` string in `config/config.yaml`/`config/dev.yaml` (currently `v0.1.88`), and a
  separate git-tag-based release step (`RedSockActions/release_tag`) in CI. `goreleaser` versions
  packages from git tags by default — adopting it means deciding whether `app_info.version` keeps
  being hand-edited or is dropped/auto-synced (see decisions callout).

## Headline recommendation

**Adopt [goreleaser](https://goreleaser.com) as the single build+release pipeline for all three
targets, replacing (or running alongside) the current custom `RedSockActions/release_image` +
MinIO-upload flow for the *binary* release.** One `.goreleaser.yml` can:

1. Cross-compile the static Go binary for `linux/amd64`, `linux/arm64`, `darwin/amd64`,
   `darwin/arm64`, and `windows/amd64` — mirroring what the Dockerfile already does with
   `GOOS`/`GOARCH`/`CGO_ENABLED=0`.
2. Build `.deb` (and, at no extra cost, `.rpm`/`.apk`) packages via its built-in **nfpm**
   integration — this *is* the "build a `.deb` with nfpm" option from the apt section below, just
   invoked through goreleaser instead of run standalone.
3. Publish a **Homebrew Formula** to a tap repo via its `brews:` config.
4. Build a **Chocolatey** `.nupkg` (nuspec + install/uninstall PowerShell) via its `chocolateys:`
   config.

This is why sections 1–3 below converge: goreleaser is presented as the *build* tool for all three,
while the *distribution* decision (self-hosted apt repo vs. PPA, tap vs. homebrew-core, community
Chocolatey vs. private feed) is still a separate, per-target choice that only the repo owner can
make (see the decisions callout).

**One important nuance discovered during research, specific to Velez being a *service* and not a
plain CLI:** goreleaser deprecated its classic Homebrew `brews:` (Formula) config in v2.10 in favor
of a newer `homebrew_casks:` config, and **casks do not support the `service do ... end` block**
that lets `brew services start velez` manage a background process — only the older, deprecated
Formula path supports that today (no removal date has been announced). Since Velez needs to run as
a long-lived background service on macOS too, **the recommendation is to deliberately use the
deprecated `brews:` (Formula) path, not the newer casks path**, until goreleaser adds service
support to casks. This is flagged again in the Homebrew section.

Because none of this exists yet, adopting goreleaser is itself an architectural decision (new build
tool, new CI stage, new artifacts) — it is a recommendation for the owner to approve, not something
this doc or a future agent should do unprompted.

---

## 1. Ubuntu / Debian (apt)

### Realistic options considered

| Option                                                                       | What it is                                                                                                                                                                           | Verdict                                                                                                                                                                                                                             |
|------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| (a) Launchpad PPA                                                            | Debian-style source upload (`debian/control`, `debian/rules`, GPG-signed `.dsc`) built by Launchpad's own builders                                                                   | Heaviest option — designed for projects that want to sit in Ubuntu's official ecosystem with proper Debian packaging conventions and a public GPG signing story. Overkill for a single-team internal service at this stage.         |
| (b1) packagecloud                                                            | Hosted apt/yum repo-as-a-service, `.deb` upload via API/CLI, handles signing and repo metadata for you                                                                               | Fastest path to a real `apt install`-able repo with no infra to run, but it's a recurring paid subscription and another external account/credential to manage.                                                                      |
| (b2) Self-hosted repo (`aptly`/`reprepro`/`apt-ftparchive` + static hosting) | You generate the `Packages`/`Release`/`InRelease` metadata yourself and serve it over plain HTTPS (S3, GitHub Pages, or the MinIO bucket the release pipeline already uploads to)    | No recurring cost, and the release pipeline **already** uploads artifacts to a MinIO/S3 bucket (`scripts/init_node.sh` today) — this is the same shape of infrastructure, just adding `.deb`s and repo metadata to the same bucket. |
| (c) Just ship a `.deb`, no repo                                              | Build with `nfpm`/`fpm`/`dpkg-deb`, publish the file on GitHub Releases / the existing S3 bucket, users run `dpkg -i velez_<ver>_amd64.deb` or `apt install ./velez_<ver>_amd64.deb` | Zero repo infrastructure, zero signing setup, works today. No automatic upgrades — users re-download and re-install for every version.                                                                                              |

### Recommendation

**Start with (c) — a plain `.deb` built via goreleaser's nfpm integration, published as a GitHub
Release / S3 asset — and graduate to (b2) a self-hosted static apt repo once there's real demand
for `apt upgrade`-driven updates.** Reasoning:

- This is a small, internally-consumed Go service, not a project trying to reach the broad Ubuntu
  user base — a Launchpad PPA's overhead (Debian source packaging conventions, GPG key management,
  Launchpad account, review latency) isn't justified yet.
- packagecloud's recurring cost is hard to justify before it's clear how many nodes will actually
  install Velez this way versus continuing to use the Docker image.
- A self-hosted repo (b2) is the natural "phase 2": the release pipeline already has S3/MinIO
  upload wired up, so adding `apt-ftparchive`-generated `Packages`/`Release` files to that same
  bucket is incremental, not a new integration.
- Because the build step is nfpm either way (standalone or via goreleaser), choosing "just ship the
  `.deb`" now doesn't foreclose "add a repo" later — the same package artifact serves both.

### Package naming and versioning

- Package name: lowercase `velez`, consistent with the Docker image name (`godverv/velez` /
  `vervstack/velez` — note these two names appear in `README.md` and `scripts/prod_init.sh.pattern`
  respectively; which org/name is canonical is itself a decision, see callout).
- Version: use the git tag goreleaser is invoked against (e.g. `v0.1.88` → Debian version
  `0.1.88`), not the hand-maintained `app_info.version` field in `config/config.yaml`. Debian
  version strings can't have a leading `v`; nfpm/goreleaser strip it automatically.
- Architecture: build `amd64` and `arm64` `.deb`s — matches the Dockerfile's existing
  `GOOS`/`GOARCH` cross-compilation and covers both typical cloud VMs (amd64) and ARM nodes (the
  Makefile already targets `build-local-container` for ARM64).

### systemd unit — needs to be authored, does not exist yet

There is no `.service` file in this repo. A unit will need to be written as part of implementation
(not by this doc) along these lines, for review:

- `ExecStart=/usr/bin/velez -config /etc/velez/config.yaml` — **the explicit `-config` flag is
  required**, since the binary's default (`./config/config.yaml`, relative to CWD) will not
  resolve correctly under systemd unless `WorkingDirectory=` is also pinned to a directory
  containing a `config/` subfolder, which is a fragile approach for a native package. Passing
  `-config` directly is the deliberate fix.
- `User=`/`Group=` and Docker socket access: the process needs to read/write
  `/var/run/docker.sock`, so the service user needs membership in the host's `docker` group (or the
  unit needs to run as root). This is a real security decision — see the callout list.
- `Restart=on-failure` and `WantedBy=multi-user.target`, matching the `--restart=always` behavior
  the current Docker-based install already relies on (`scripts/prod_init.sh.pattern`).
- nfpm can embed the unit file directly in the `.deb` (via its `contents:` list) and enable it in a
  postinstall script — see below.

### Config file placement

- Ship a default config at **`/etc/velez/config.yaml`**, following the standard `/etc/<pkg>/`
  convention — this replaces the Docker image's `/app/config/config.yaml`.
- The current containerized default for local state (`local_state_path`,
  e.g. `/tmp/velez/local_state.json` in `config.yaml`) is **not appropriate for a native
  install** — `/tmp` is ephemeral and world-writable-adjacent on most distros. Recommend
  `/var/lib/velez/` for persisted state (private key, etc.), matching the FHS convention for
  service-owned mutable state, and overriding it in the shipped `/etc/velez/config.yaml`.
- Don't overwrite an existing `/etc/velez/config.yaml` on upgrade — nfpm marks config files as
  `type: config|noreplace` in its file list for exactly this reason; confirm this is set correctly
  when the package is actually built.

### postinst/prerm scripts

Because a systemd service needs to be enabled/started (and cleanly stopped on removal), the `.deb`
needs standard maintainer scripts — nfpm supports specifying these directly:

- `postinstall`: `systemctl daemon-reload && systemctl enable --now velez.service` (skip
  `--now` if the owner prefers packages to not auto-start on install).
- `preremove`: `systemctl disable --now velez.service` (or `stop` only, on upgrade — nfpm/dpkg
  distinguish install-time vs. upgrade-time removal via the same script, the logic needs the usual
  `$1` argument check dpkg maintainer scripts use).
- `postremove`: `systemctl daemon-reload` after a full purge.

### Step-by-step (once goreleaser is adopted)

1. Add a `nfpms:` block to `.goreleaser.yml` targeting `deb` (and optionally `rpm`, `apk`) formats,
   pointing at the compiled `linux/amd64` and `linux/arm64` binaries.
2. Add `contents:` entries for the systemd unit (→ `/lib/systemd/system/velez.service`) and default
   config (→ `/etc/velez/config.yaml`, marked `noreplace`).
3. Add `scripts.postinstall` / `scripts.preremove` pointing at small shell scripts checked into the
   repo (e.g. `packaging/deb/postinstall.sh`) — not written yet, future step.
4. Run `goreleaser release` (or `goreleaser build --snapshot` locally to test) — this produces the
   `.deb` artifacts alongside the existing Docker image build.
5. Publish the `.deb` as a GitHub Release asset (simplest) and/or push it into a self-hosted apt
   repo once that's set up (`aptly publish` or `apt-ftparchive` + sync to the S3 bucket).

---

## 2. Homebrew (macOS)

### Tap vs. homebrew-core

- **A tap (`godverv/homebrew-velez` or similar — own repo, name TBD)** is the right choice at this
  stage. `homebrew-core` has real acceptance bars (notability, stable release history, no
  significant known bugs, must build from source rather than fetch a prebuilt binary in most
  cases) that a young internal service tool won't clear, and going through core review adds
  latency for every release. A tap is entirely self-controlled: create it, push a Formula, done.
- Repo naming convention Homebrew enforces: the tap repo **must** be named `homebrew-<name>` (e.g.
  `homebrew-velez`) for `brew tap <org>/velez` to resolve it. This is a hard requirement, not a
  style choice.

### Formula structure and the goreleaser nuance

- Use `goreleaser`'s Homebrew integration to generate and push the Formula on every release,
  rather than hand-maintaining one. It downloads the correct prebuilt `darwin/amd64` or
  `darwin/arm64` archive from the release and writes the Formula's `install` block automatically.
- **Use the classic `brews:` config (Formula), not the newer `homebrew_casks:` config** — this is a
  deliberate, non-default choice. goreleaser deprecated `brews:` in v2.10 in favor of
  `homebrew_casks:`, but casks currently have no equivalent to a Formula's `service do ... end`
  block, so `brew services start/stop/restart velez` — the standard way macOS users manage a
  background service installed via Homebrew — isn't available for casks yet. Since Velez is a
  long-running server, not a one-shot CLI, the Formula path is the only one of the two that gives
  users normal service management today. Revisit this if/when goreleaser adds cask service support
  (there's an open discussion upstream but no shipped feature or date as of this writing).
- The Formula's `service do ... end` block is Homebrew's declarative equivalent of a systemd unit /
  `.plist` — goreleaser/Homebrew generate the actual launchd `.plist` from it, so unlike section 1
  no unit file needs to be hand-authored for macOS. It typically needs: the run command
  (`opt_bin/"velez" "-config" "#{etc}/velez/config.yaml"` — again, the explicit `-config` flag
  matters here for the same reason as systemd), a working directory, and log file paths
  (`var/log/velez.log`).
- Config placement under Homebrew's prefix convention: `#{etc}/velez/config.yaml` (i.e.
  `/opt/homebrew/etc/velez/config.yaml` on Apple Silicon, `/usr/local/etc/velez/config.yaml` on
  Intel) — the Formula's `install` step should copy a default config there if one doesn't already
  exist, without clobbering it on upgrade.

### Step-by-step (once goreleaser is adopted)

1. Create the `homebrew-velez` tap repository under whichever GitHub org owns releases.
2. Add a `brews:` block to `.goreleaser.yml`: `repository:` pointing at the tap, `homepage`,
   `description`, and a `service` block per above.
3. Give goreleaser a token with push access to the tap repo (separate `GITHUB_TOKEN`/PAT secret
   from whatever the Docker release pipeline uses).
4. On release, goreleaser pushes an updated `velez.rb` Formula to the tap automatically.
5. Users install with `brew tap <org>/velez && brew install velez && brew services start velez`.

### Docker's host-access requirements don't disappear on macOS

Same caveat as section 1: the Formula-installed binary still needs to reach a Docker daemon (e.g.
Docker Desktop's socket on macOS) — this needs to be documented in the Formula's `caveats` block so
`brew install` tells the user what to configure, since macOS doesn't have `/var/run/docker.sock` by
default the way Linux does.

---

## 3. Chocolatey (Windows)

### Does Velez support running as a native Windows service today?

**No.** Confirmed by checking `go.mod` (no `github.com/kardianos/service`, no
`golang.org/x/sys/windows/svc` import — `golang.org/x/sys` is present only as an indirect
dependency of the Docker SDK) and `cmd/service/main.go` (a plain `main()` that runs in the
foreground with no Windows service-control-manager integration). This is the same situation as
Linux/macOS: the binary itself doesn't know how to be a service, so whatever wraps it has to
provide that.

### Two ways to make it a real Windows service, and the recommendation

1. **Wrap the existing binary with NSSM (or `sc.exe create` + a thin batch/PowerShell shim) from
   the Chocolatey install script.** Zero Go code changes — the choco package's
   `chocolateyinstall.ps1` downloads/installs NSSM (or uses `sc.exe`, which ships with Windows) and
   registers the plain `velez.exe -config C:\ProgramData\velez\config.yaml` as a service. This is
   the pragmatic near-term option and mirrors exactly what sections 1 and 2 do — an OS-level
   supervisor wrapping an unmodified binary — so it keeps all three platforms conceptually
   consistent without touching application code.
2. **Add a real Windows service entrypoint to Velez** (`golang.org/x/sys/windows/svc` or
   `kardianos/service`, the latter also smooths over the Linux/macOS service-detection code path).
   More "native" — `services.msc` sees Velez as a normal service without an NSSM layer in between —
   but it's a Go code change to `cmd/service/main.go`, not just a packaging change, and should be
   scoped and approved separately from this packaging doc.

**Recommendation: start with option 1 (NSSM-wrapped) to unblock a Chocolatey package without
touching Go code**, and treat option 2 as a possible follow-up if the NSSM layer proves annoying
operationally (it adds one more moving part users have to reason about when debugging service
issues).

### Package structure

- `.nuspec` — package metadata (id `velez`, version, description, dependencies). goreleaser has
  native `chocolateys:` support that generates the `.nupkg` (nuspec + tools folder) directly from
  the release build, so this doesn't need to be hand-maintained either.
- `tools\chocolateyinstall.ps1` — downloads (or references a bundled) `velez.exe`, lays down a
  default config at e.g. `C:\ProgramData\velez\config.yaml`, and registers the Windows service
  (NSSM or `sc.exe`, per the decision above).
- `tools\chocolateyuninstall.ps1` — stops and removes the service, leaves config/data in place
  unless the user explicitly asks for a full purge (standard Windows convention — don't delete
  user data silently on uninstall).

### Community repo vs. self-hosted feed

- **Community Chocolatey repository** (`community.chocolatey.org`, `choco push` with an API key):
  free, gives users a familiar `choco install velez`, but every package version goes through
  Chocolatey's automated + potentially moderator review before it's live, which adds release
  latency and is a public listing (visible to anyone).
- **Self-hosted feed** (a private NuGet-compatible feed, or Chocolatey's own self-hosted
  server/Azure DevOps artifacts/a plain file share): private, no review latency, but users need to
  be told to add the custom source (`choco source add`) before `choco install` works — one more
  manual step compared to the community repo.
- Given Velez is presumably installed on machines the team controls rather than by end users
  browsing a public package index, **a self-hosted feed is the more consistent choice with the
  self-hosted-apt-repo recommendation in section 1** — but this is genuinely a coin-flip decision
  that depends on who's expected to install Velez on Windows, so it's listed in the decisions
  callout rather than being decided here.

### Step-by-step (once goreleaser is adopted)

1. Decide NSSM vs. `sc.exe` vs. native Go service support (see above).
2. Add a `chocolateys:` block to `.goreleaser.yml` with the nuspec metadata and pointers to the
   install/uninstall PowerShell scripts (checked into e.g. `packaging/choco/`, not written yet).
3. Get a Chocolatey API key (community) or stand up a private feed, and give goreleaser the
   corresponding push credentials.
4. On release, goreleaser builds and pushes the `.nupkg`.
5. Users install with `choco install velez` (community) or
   `choco install velez --source=<private-feed-url>` (self-hosted).

---

## Prerequisites / decisions needed from you

These are things only the repo owner can decide — nothing above should be treated as settled until
these are answered:

1. **Adopt goreleaser at all?** This is the biggest one — it's a new build tool and a new CI stage
   alongside (or replacing parts of) the existing `RedSockActions/release_image` pipeline. Worth
   confirming before any `.goreleaser.yml` gets written.
2. **Canonical package/binary name and org.** `README.md` uses `godverv/velez`,
   `scripts/prod_init.sh.pattern` uses `vervstack/velez` — these two already disagree. Pick one
   name to carry into the `.deb` package name, the Homebrew tap (`<org>/homebrew-velez`), and the
   Chocolatey package id, and confirm `velez` isn't already taken in each namespace (Chocolatey
   community repo and Homebrew core/other taps, in particular — worth a quick manual check before
   committing to it).
3. **Self-host vs. use community/public repos**, independently for each target: self-hosted apt
   repo vs. packagecloud vs. GitHub-Releases-only for apt; a tap (already recommended) but which
   GitHub org owns it; community Chocolatey repo vs. a private feed. These have different ongoing
   maintenance and cost profiles.
4. **Does the native install need to run as root, or is `docker`-group membership + reduced
   hardware-discovery capability acceptable?** This affects the systemd `User=`/Chocolatey service
   account/Homebrew Formula `service` block on every platform. Least-privilege is preferable but
   may reduce functionality (`ghw` hardware discovery in particular may need more than group
   membership on some distros) — needs an explicit call, not a default.
5. **Signing.** apt repos are conventionally GPG-signed (`Release`/`InRelease` signing key that
   `apt` verifies); Homebrew taps don't require this since GitHub itself is the trust anchor;
   Chocolatey packages can optionally be code-signed. Decide whether to generate/manage a GPG
   signing key for apt now or ship unsigned initially (with users adding
   `[trusted=yes]`/`--allow-unauthenticated`, which is a real security downgrade worth being
   explicit about rather than silently accepting).
6. **Windows service strategy** — NSSM/`sc.exe` wrapper (recommended near-term, zero Go changes) vs.
   adding real Windows-service support to `cmd/service/main.go` (bigger change, more "native"
   result, would also let Linux/macOS share a service-detection library like `kardianos/service` if
   desired).
7. **Version source of truth.** Keep `app_info.version` in `config/config.yaml`/`dev.yaml` as a
   separately hand-maintained value, or drop/auto-sync it once goreleaser (which versions from git
   tags) is doing the packaging? Divergence between the two today (e.g. package `v0.2.0` shipping
   with `app_info.version: v0.1.88` baked into the default config) would be confusing.
8. **State/config directory paths.** This doc proposes `/etc/velez/config.yaml` +
   `/var/lib/velez/` (Linux), `#{etc}/velez/` (Homebrew), `C:\ProgramData\velez\` (Windows) as FHS/
   platform-idiomatic replacements for the current container-only `/tmp/velez/...` default — confirm
   these are acceptable before they're baked into postinstall scripts, since changing them later
   means a migration for anyone who already installed a native package.
