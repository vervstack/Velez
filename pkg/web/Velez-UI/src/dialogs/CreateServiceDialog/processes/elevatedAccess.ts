import type {DockerContainer, ElevatedAccess} from "@/app/api/velez"

export interface ElevatedAccessState {
    isElevated: boolean
    isAllowRequired: boolean
    isBlocked: boolean
}

export function isElevatedAccessPresent(access?: ElevatedAccess): boolean {
    if (!access) return false
    return Boolean(
        access.isPrivileged
        || access.isHostPid
        || access.isHostIpc
        || access.isHostUserns
        || (access.capAdd ?? []).length > 0
        || (access.devices ?? []).length > 0,
    )
}

export function elevatedAccessStateOf(
    container: DockerContainer,
    isRecreating: boolean,
    isAllowed: boolean,
): ElevatedAccessState {
    const isElevated = isRecreating && isElevatedAccessPresent(container.elevatedAccess)
    const isAllowRequired = isElevated && !container.isImageWhitelisted
    return {isElevated, isAllowRequired, isBlocked: isAllowRequired && !isAllowed}
}

export function elevatedAccessItemsOf(access: ElevatedAccess): string[] {
    return [
        ...(access.isPrivileged ? ["Privileged"] : []),
        ...(access.devices ?? []).map((device) => `Device ${device.hostPath ?? ""}`),
        ...(access.capAdd ?? []).map((capability) => `Cap ${capability}`),
        ...(access.isHostPid ? ["Host PID"] : []),
        ...(access.isHostIpc ? ["Host IPC"] : []),
        ...(access.isHostUserns ? ["Host user namespace"] : []),
    ]
}
