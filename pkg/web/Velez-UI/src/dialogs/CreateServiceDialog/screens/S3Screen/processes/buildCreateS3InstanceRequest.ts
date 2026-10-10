import type {CreateS3InstanceRequest} from "@/app/api/velez/s3_api.pb"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"

export const LOCKED_REPLICATION_FACTOR = 1

export interface CreateS3InstanceFormState {
    name: string
    environment: string
    box: string
    isPortExposed: boolean
    port: string
    region: string
    isWebUiEnabled: boolean
}

export function buildCreateS3InstanceRequest(form: CreateS3InstanceFormState): CreateS3InstanceRequest | null {
    const trimmedName = form.name.trim()
    if (!trimmedName || validateInstanceName(trimmedName)) return null

    const trimmedPort = form.port.trim()

    return {
        name: trimmedName,
        environment: form.environment || undefined,
        box: form.box,
        exposeToPort: form.isPortExposed && trimmedPort ? Number(trimmedPort) : undefined,
        replicationFactor: LOCKED_REPLICATION_FACTOR,
        region: form.region.trim() || undefined,
        enableWebUi: form.isWebUiEnabled,
    }
}
