import type {DockerContainer, RegisterContainerRequest} from "@/app/api/velez"
import {
    initialRoles,
    isGenericOnly,
    RegisterRoles,
    serviceIdOf,
} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"
import {ADOPTABLE_SCREENS, ProductScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

export interface OnboardFormState {
    request: RegisterContainerRequest | null
    isRecreating: boolean
    hasPublishedPorts: boolean
    blockReason?: string
}

export interface OnboardContext {
    container: DockerContainer
    containers: DockerContainer[]
    roles: RegisterRoles
    serviceNameDraft?: string
    screen?: ProductScreen
    form?: OnboardFormState
}

export function initialOnboardContext(container: DockerContainer, networkMembers: DockerContainer[]): OnboardContext {
    return {
        container,
        containers: [container, ...networkMembers],
        roles: initialRoles(container.id ?? ""),
    }
}

export function nameOfContainer(context: OnboardContext, containerId: string): string {
    return context.containers.find((member) => member.id === containerId)?.name ?? ""
}

export function serviceNameOf(context: OnboardContext): string {
    return context.serviceNameDraft ?? nameOfContainer(context, serviceIdOf(context.roles))
}

export function isRootChanged(context: OnboardContext): boolean {
    return context.roles.rootId !== context.container.id
}

export function isRolesChanged(context: OnboardContext): boolean {
    return isGenericOnly(context.roles, context.container.id ?? "")
}

export function enabledScreensOf(context: OnboardContext): ProductScreen[] {
    return isRolesChanged(context) ? ["generic"] : ADOPTABLE_SCREENS
}

export function selectedScreenOf(context: OnboardContext): ProductScreen | undefined {
    const enabled = enabledScreensOf(context)
    if (context.screen !== undefined && enabled.includes(context.screen)) return context.screen
    if (enabled.length === 1) return enabled[0]
    return undefined
}

export function isTypeStepReady(context: OnboardContext): boolean {
    return selectedScreenOf(context) !== undefined
}

export function networkModeOf(container: DockerContainer, context: OnboardContext): string {
    const ownerId = container.networkOwnerContainerId
    if (ownerId) {
        const ownerName = nameOfContainer(context, ownerId) || context.container.name
        return `container:${ownerName ?? ""}`
    }

    const names = (container.networks ?? []).map((network) => network.networkName).filter(Boolean)
    return names.length > 0 ? names.join(", ") : "default"
}
