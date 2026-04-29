# Velez — Technical Review Index

Velez is a Docker-based deployment platform at an early but functional stage. The core create/list/drop container loop
works. The higher-level abstractions (Services, Deployments, upgrade lifecycle, config subscriptions) are wired up
structurally but are largely stubs — the worker that drives them has a logic bug that prevents most of it from running
at all. Networking (Headscale/VCN) and multi-node clustering are partially integrated but disabled at key paths.

The goal of this review is to define the work needed to get from "structurally shaped" to "actually works."

## Documents

| File                               | Contents                                                                                                         |
|------------------------------------|------------------------------------------------------------------------------------------------------------------|
| [bugs.md](bugs.md)                 | Confirmed bugs — wrong behavior, panics, silent failures. All have file:line references.                         |
| [incomplete.md](incomplete.md)     | Features that are wired but stubbed: empty switch cases, nil returns, commented-out code, empty proto responses. |
| [architecture.md](architecture.md) | Structural design concerns that don't manifest as bugs today but will block adding podman/k8s/multi-node later.  |
| [roadmap.md](roadmap.md)           | Milestones from "unfuck current setup" through podman, k8s, and clustering.                                      |
