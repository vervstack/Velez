import type {CreateRunnerRequest, RunnerScope} from "@/app/api/velez"

interface CreateRunnerFormState {
    name: string
    scope: RunnerScope
    target: string
    labels: string
    environment: string
    accessToken: string
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

    return {
        name: trimmedName,
        scope: form.scope,
        target: trimmedTarget,
        labels: parsedLabels,
        environment: form.environment || undefined,
        dockerSocketAddress: form.dockerSocketAddress || undefined,
        github: {
            accessToken: trimmedAccessToken,
        },
    }
}
