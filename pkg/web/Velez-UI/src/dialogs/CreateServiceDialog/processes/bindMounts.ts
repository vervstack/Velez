import type {DockerContainer} from "@/app/api/velez"

export interface BindMount {
    source: string
    destination: string
}

export interface ResolvedLink extends BindMount {
    volumeName: string
}

const INVALID_VOLUME_NAME_CHARS = /[^a-zA-Z0-9_.-]+/g
const SURROUNDING_SLASHES = /^\/+|\/+$/g

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
): ResolvedLink[] {
    return mounts.map((mount) => ({
        ...mount,
        volumeName: overridesByDestination[mount.destination]
            ?? defaultLinkVolumeName(serviceName, mount.destination),
    }))
}
