import type {CreateDindRequest} from "@/app/api/velez/dind_api.pb"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"

interface CreateDindFormState {
    name: string
    environment: string
    isSysboxEnabled: boolean
}

export function buildCreateDindRequest(form: CreateDindFormState): CreateDindRequest | null {
    const trimmedName = form.name.trim()
    if (!trimmedName || validateInstanceName(trimmedName)) return null

    return {
        name: trimmedName,
        environment: form.environment.trim() || undefined,
        isSysboxEnabled: form.isSysboxEnabled,
    }
}
