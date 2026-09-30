import {RunnerProvider, RunnerScope, type DockerContainer, type ServicePattern} from "@/app/api/velez"
import type {RunnerForm} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"
import {runnerProviderOf, suggestedScreenOf} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

export function parseRunnerProvider(value?: string): RunnerProvider | undefined {
    if (value === RunnerProvider.GITHUB) return RunnerProvider.GITHUB
    if (value === RunnerProvider.GITLAB) return RunnerProvider.GITLAB
    return undefined
}

export function parseRunnerScope(value?: string): RunnerScope | undefined {
    if (value === RunnerScope.REPO) return RunnerScope.REPO
    if (value === RunnerScope.ORG) return RunnerScope.ORG
    return undefined
}

function providerOfImage(imageName?: string): RunnerProvider | undefined {
    if (!imageName) return undefined
    if (/gitlab/i.test(imageName)) return RunnerProvider.GITLAB
    if (/github/i.test(imageName)) return RunnerProvider.GITHUB
    return undefined
}

function providerOfPattern(pattern?: ServicePattern): RunnerProvider | undefined {
    const screen = suggestedScreenOf(pattern)
    return screen ? runnerProviderOf(screen) : undefined
}

export function initialRunnerForm(container: DockerContainer): RunnerForm {
    const defaults = container.suggestedRunnerDefaults
    const provider = parseRunnerProvider(defaults?.provider) ??
        providerOfPattern(container.suggestedPattern) ??
        providerOfImage(container.imageName)

    return {
        provider,
        scope: parseRunnerScope(defaults?.scope) ?? RunnerScope.REPO,
        target: defaults?.target ?? "",
        baseUrl: defaults?.baseUrl ?? "",
        labels: (defaults?.labels ?? []).join(", "),
        dockerImage: "",
        concurrent: "",
        accessToken: "",
        registrationToken: "",
    }
}
