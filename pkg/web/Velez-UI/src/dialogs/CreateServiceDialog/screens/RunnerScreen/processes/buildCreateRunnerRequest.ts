import {RunnerProvider, type CreateRunnerRequest, type GitlabConfig, type RunnerScope} from "@/app/api/velez"
import {parseConcurrent} from "@/processes/parseConcurrent.ts"

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
    dindName: string
    dockerSocketAddress: string
    concurrent: string
}

export function buildCreateRunnerRequest(form: CreateRunnerFormState): CreateRunnerRequest | null {
    const trimmedName = form.name.trim()
    if (!trimmedName) return null

    const trimmedTarget = form.target.trim()
    if (!trimmedTarget) return null

    const trimmedAccessToken = form.accessToken.trim()
    if (!trimmedAccessToken) return null

    if (!form.dindName === !form.dockerSocketAddress) return null

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
        dindName: form.dindName || undefined,
        dockerSocketAddress: form.dockerSocketAddress || undefined,
    }

    if (form.provider === RunnerProvider.GITLAB) {
        const gitlabConfig: GitlabConfig = {
            accessToken: trimmedAccessToken,
            baseUrl: form.baseUrl.trim() || undefined,
            dockerImage: form.dockerImage.trim() || undefined,
            concurrent: parseConcurrent(form.concurrent),
        }

        return {
            ...base,
            gitlab: gitlabConfig,
        }
    }

    return {
        ...base,
        github: {accessToken: trimmedAccessToken},
    }
}
