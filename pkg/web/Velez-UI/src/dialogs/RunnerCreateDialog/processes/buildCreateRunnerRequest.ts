import {RunnerProvider, type CreateRunnerRequest, type RunnerScope} from "@/app/api/velez"

interface CreateRunnerFormState {
    name: string
    provider: RunnerProvider
    scope: RunnerScope
    target: string
    labels: string
    environment: string
    accessToken: string
    baseUrl: string
    dockerImage: string
    dockerSocketAddress: string
}

export function buildCreateRunnerRequest(form: CreateRunnerFormState): CreateRunnerRequest | null {
    const trimmedName = form.name.trim()
    if (!trimmedName) return null

    const trimmedTarget = form.target.trim()
    if (!trimmedTarget) return null

    const trimmedAccessToken = form.accessToken.trim()
    if (!trimmedAccessToken) return null

    const parsedLabels = form.labels
        .split(",")
        .map(l => l.trim())
        .filter(l => l.length > 0)

    const base = {
        name: trimmedName,
        scope: form.scope,
        target: trimmedTarget,
        labels: parsedLabels,
        environment: form.environment || undefined,
        dockerSocketAddress: form.dockerSocketAddress || undefined,
    }

    if (form.provider === RunnerProvider.GITLAB) {
        return {
            ...base,
            gitlab: {
                accessToken: trimmedAccessToken,
                baseUrl: form.baseUrl.trim() || undefined,
                dockerImage: form.dockerImage.trim() || undefined,
            },
        }
    }

    return {
        ...base,
        github: {accessToken: trimmedAccessToken},
    }
}
