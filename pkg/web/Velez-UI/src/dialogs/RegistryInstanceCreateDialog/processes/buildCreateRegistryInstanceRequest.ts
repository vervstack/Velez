import type {CreateRegistryInstanceRequest} from "@/app/api/velez"

interface CreateRegistryInstanceFormState {
    name: string
    environment: string
    box: string
    exposePort: boolean
    port: string
    ownerService: string
}

export function buildCreateRegistryInstanceRequest(
    form: CreateRegistryInstanceFormState
): CreateRegistryInstanceRequest | null {
    const trimmedName = form.name.trim()
    if (!trimmedName) return null

    return {
        name: trimmedName,
        environment: form.environment || undefined,
        box: form.box,
        exposeToPort: form.exposePort && form.port.trim() ? Number(form.port.trim()) : undefined,
        ownerService: form.ownerService || undefined,
    }
}
