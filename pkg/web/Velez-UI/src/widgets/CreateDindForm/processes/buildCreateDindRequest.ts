import type {CreateDindRequest} from "@/app/api/velez/dind_api.pb"

interface CreateDindFormState {
    name: string
    environment: string
    isSysboxEnabled: boolean
}

export function buildCreateDindRequest(form: CreateDindFormState): CreateDindRequest | null {
    const trimmedName = form.name.trim()
    if (!trimmedName) return null

    return {
        name: trimmedName,
        environment: form.environment.trim() || undefined,
        isSysboxEnabled: form.isSysboxEnabled,
    }
}
