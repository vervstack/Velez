import type {DockerContainer} from "@/app/api/velez"

export interface BindMount {
    source: string
    destination: string
}

export interface ResolvedLink extends BindMount {
    volumeName: string
    isKeptAsBind: boolean
}

const INVALID_VOLUME_NAME_CHARS = /[^a-zA-Z0-9_.-]+/g
const SURROUNDING_SLASHES = /^\/+|\/+$/g
const SYSTEM_PATH_ROOTS = ["/dev", "/proc", "/sys", "/run", "/var/run"]
const SOCKET_SUFFIX = ".sock"

export function isSystemPath(source: string): boolean {
    if (source.endsWith(SOCKET_SUFFIX)) return true
    return SYSTEM_PATH_ROOTS.some((root) => source === root || source.startsWith(`${root}/`))
}

export function bindMountsOf(container: DockerContainer): BindMount[] {
    return (container.mounts ?? [])
        .filter((mount) => mount.type === "bind" && mount.source && mount.destination)
        .map((mount) => ({source: mount.source ?? "", destination: mount.destination ?? ""}))
}

// Mirrors deriveLinkedVolumeName in internal/jobs/register_container_links.go.
export function defaultLinkVolumeName(serviceName: string, target: string): string {
    const trimmedTarget = target.replace(SURROUNDING_SLASHES, "")
    return `${serviceName.trim()}_${trimmedTarget}`.replace(INVALID_VOLUME_NAME_CHARS, "_")
}

export function resolveLinks(
    mounts: BindMount[],
    serviceName: string,
    overridesByDestination: Record<string, string>,
    keptOverridesByDestination: Record<string, boolean>,
): ResolvedLink[] {
    return mounts.map((mount) => ({
        ...mount,
        volumeName: overridesByDestination[mount.destination]
            ?? defaultLinkVolumeName(serviceName, mount.destination),
        isKeptAsBind: keptOverridesByDestination[mount.destination] ?? isSystemPath(mount.source),
    }))
}
