import type {CreatePgInstanceRequest} from "@/app/api/velez"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"

interface CreatePgInstanceFormState {
    name: string
    environment: string
    box: string
    exposePort: boolean
    port: string
    ownerService: string
}

export function buildCreatePgInstanceRequest(form: CreatePgInstanceFormState): CreatePgInstanceRequest | null {
    const trimmedName = form.name.trim()
    if (!trimmedName || validateInstanceName(trimmedName)) return null

    return {
        name: trimmedName,
        environment: form.environment || undefined,
        box: form.box,
        exposeToPort: form.exposePort && form.port.trim() ? Number(form.port.trim()) : undefined,
        ownerService: form.ownerService || undefined,
    }
}
