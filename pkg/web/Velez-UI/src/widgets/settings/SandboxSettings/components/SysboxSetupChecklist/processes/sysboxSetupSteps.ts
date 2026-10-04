export type SysboxStepId =
    | "linux"
    | "kernel"
    | "docker"
    | "install"
    | "services"
    | "runtime"
    | "smoke"
    | "enable"
    | "dind"

export interface SysboxSetupStep {
    id: SysboxStepId
    title: string
    detail: string
    command?: string
    expected?: string
    fallback?: string
}

export const SYSBOX_CHECKLIST_INTRO =
    "Sysbox must be installed on the host that runs Docker before the toggle can be enabled. " +
    "Velez refuses to enable it while the sysbox-runc runtime is not registered."

export const SYSBOX_SETUP_STEPS: SysboxSetupStep[] = [
    {
        id: "linux",
        title: "Linux host only",
        detail: "Sysbox does not run on macOS, Windows or Docker Desktop. " +
            "Use a Linux VM (Ubuntu/Debian are the best-supported) with systemd.",
    },
    {
        id: "kernel",
        title: "Recent kernel",
        detail: "Kernel 5.12 or newer is recommended (ID-mapped mounts); older kernels need shiftfs.",
    },
    {
        id: "docker",
        title: "Regular Docker Engine",
        detail: "Install Docker Engine from the official packages. " +
            "Snap-packaged and rootless Docker are not supported.",
    },
    {
        id: "install",
        title: "Install Sysbox",
        detail: "Install the Sysbox package for your distro and architecture from the Sysbox releases page " +
            "(github.com/nestybox/sysbox). The installer registers the runtime and restarts Docker — " +
            "containers without a restart policy will be stopped, so run it in a maintenance window.",
    },
    {
        id: "services",
        title: "Check the services",
        detail: "All three Sysbox services must be running.",
        command: "systemctl status sysbox sysbox-mgr sysbox-fs",
        expected: "active (running)",
    },
    {
        id: "runtime",
        title: "Check Docker sees the runtime",
        detail: "The runtimes list must contain sysbox-runc.",
        command: "docker info | grep -i runtimes",
        expected: "sysbox-runc",
        fallback: "If missing, add the following to /etc/docker/daemon.json and restart Docker:\n" +
            "{\"runtimes\": {\"sysbox-runc\": {\"path\": \"/usr/bin/sysbox-runc\"}}}",
    },
    {
        id: "smoke",
        title: "Smoke test",
        detail: "A container started under Sysbox prints ok.",
        command: "docker run --rm --runtime=sysbox-runc alpine echo ok",
    },
    {
        id: "enable",
        title: "Enable it here",
        detail: "Turn on \"Run containers in Sysbox\" above. It applies to containers created afterwards; " +
            "recreate existing ones to move them under Sysbox. Privileged system containers (VPN sidecar, runners) " +
            "stay on the default runtime unless \"Apply Sysbox to whitelisted containers too\" is on.",
    },
    {
        id: "dind",
        title: "Per-DinD note",
        detail: "A Docker daemon created with Sysbox isolation also needs this setup; " +
            "without Sysbox it runs privileged.",
    },
]
