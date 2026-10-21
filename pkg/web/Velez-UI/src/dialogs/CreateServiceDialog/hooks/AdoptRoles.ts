import {useState} from "react"

import type {DockerContainer} from "@/app/api/velez"
import type {AdoptGroup} from "@/dialogs/CreateServiceDialog/hooks/AdoptGroup.ts"
import {
    initialRoles,
    isGenericOnly,
    pickRoot,
    pickService,
    RegisterRoles,
    serviceIdOf,
} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

export interface AdoptRoles {
    group: AdoptGroup
    isRolesChanged: boolean
    pendingRootId: string | null
    oldRootName: string
    pendingRootName: string
    onPickService(containerId: string): void
    onRequestRoot(containerId: string): void
    onConfirmRoot(): void
    onCancelRoot(): void
}

export function useAdoptRoles(container: DockerContainer | undefined, networkMembers: DockerContainer[]): AdoptRoles {
    const originalRootId = container?.id ?? ""
    const [roles, setRoles] = useState<RegisterRoles>(() => initialRoles(originalRootId))
    const [serviceNameDraft, setServiceNameDraft] = useState<string | undefined>(undefined)
    const [pendingRootId, setPendingRootId] = useState<string | null>(null)

    const containers = container ? [container, ...networkMembers] : []

    function nameOf(containerId: string | null): string {
        return containers.find((member) => member.id === containerId)?.name ?? ""
    }

    function handleConfirmRoot() {
        if (pendingRootId !== null) {
            setRoles(pickRoot(roles, pendingRootId, containers.map((member) => member.id ?? "")))
        }
        setPendingRootId(null)
    }

    function handlePickService(containerId: string) {
        setRoles(pickService(roles, containerId))
    }

    function handleCancelRoot() {
        setPendingRootId(null)
    }

    return {
        group: {
            containers,
            originalRootId,
            roles,
            serviceName: serviceNameDraft ?? nameOf(serviceIdOf(roles)),
            onServiceNameChange: setServiceNameDraft,
        },
        isRolesChanged: container ? isGenericOnly(roles, originalRootId) : false,
        pendingRootId,
        oldRootName: nameOf(roles.rootId),
        pendingRootName: nameOf(pendingRootId),
        onPickService: handlePickService,
        onRequestRoot: setPendingRootId,
        onConfirmRoot: handleConfirmRoot,
        onCancelRoot: handleCancelRoot,
    }
}
