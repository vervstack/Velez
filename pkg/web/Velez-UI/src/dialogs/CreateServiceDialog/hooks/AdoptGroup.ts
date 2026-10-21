import {createContext, useContext} from "react"

import type {DockerContainer} from "@/app/api/velez"
import {RegisterRoles} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

export interface AdoptGroup {
    containers: DockerContainer[]
    originalRootId: string
    roles: RegisterRoles
    serviceName: string
    blockReason?: string
    onServiceNameChange(name: string): void
}

export const AdoptGroupContext = createContext<AdoptGroup | null>(null)

export function useAdoptGroup(): AdoptGroup | null {
    return useContext(AdoptGroupContext)
}
