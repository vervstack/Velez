import type {CreateRegistryInstanceRequest} from "@/app/api/velez"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"
import {
    buildRegistryS3Storage,
    RegistryStorageState,
} from "@/dialogs/CreateServiceDialog/screens/RegistryScreen/processes/registryStorage.ts"

interface CreateRegistryInstanceFormState {
    name: string
    environment: string
    box: string
    exposePort: boolean
    port: string
    ownerService: string
    enableUi: boolean
    storage: RegistryStorageState
}

export function buildCreateRegistryInstanceRequest(
    form: CreateRegistryInstanceFormState
): CreateRegistryInstanceRequest | null {
    const trimmedName = form.name.trim()
    if (!trimmedName || validateInstanceName(trimmedName)) return null

    return {
        name: trimmedName,
        environment: form.environment || undefined,
        box: form.box,
        exposeToPort: form.exposePort && form.port.trim() ? Number(form.port.trim()) : undefined,
        ownerService: form.ownerService || undefined,
        enableUi: form.enableUi,
        s3Storage: buildRegistryS3Storage(form.storage),
    }
}
