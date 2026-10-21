export interface RegisterRoles {
    rootId: string
    explicitServiceId?: string
}

export interface RoleIds {
    rootContainerId?: string
    serviceContainerId?: string
}

export function initialRoles(originalRootId: string): RegisterRoles {
    return {rootId: originalRootId}
}

export function serviceIdOf(roles: RegisterRoles): string {
    return roles.explicitServiceId ?? roles.rootId
}

export function pickService(roles: RegisterRoles, containerId: string): RegisterRoles {
    if (containerId === roles.rootId) return {rootId: roles.rootId}
    return {rootId: roles.rootId, explicitServiceId: containerId}
}

export function pickRoot(roles: RegisterRoles, containerId: string, memberIds: string[]): RegisterRoles {
    const service = roles.explicitServiceId
    const isServiceKept = service !== undefined && service !== containerId && memberIds.includes(service)
    if (!isServiceKept) return {rootId: containerId}
    return {rootId: containerId, explicitServiceId: service}
}

export function isGenericOnly(roles: RegisterRoles, originalRootId: string): boolean {
    return roles.rootId !== originalRootId || serviceIdOf(roles) !== originalRootId
}

export function roleIdsOf(roles: RegisterRoles, originalRootId: string): RoleIds {
    const ids: RoleIds = {}
    if (roles.rootId !== originalRootId) ids.rootContainerId = roles.rootId
    if (serviceIdOf(roles) !== roles.rootId) ids.serviceContainerId = serviceIdOf(roles)
    return ids
}
