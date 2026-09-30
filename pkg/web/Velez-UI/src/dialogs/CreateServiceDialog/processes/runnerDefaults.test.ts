import {describe, expect, it} from "vitest"

import {RunnerProvider, RunnerScope, ServicePattern} from "@/app/api/velez"
import {
    initialRunnerForm,
    parseRunnerProvider,
    parseRunnerScope,
} from "@/dialogs/CreateServiceDialog/processes/runnerDefaults.ts"

describe("runnerDefaults", () => {
    it("parses provider and scope enum names and rejects anything else", () => {
        expect(parseRunnerProvider("GITHUB")).toBe(RunnerProvider.GITHUB)
        expect(parseRunnerProvider("GITLAB")).toBe(RunnerProvider.GITLAB)
        expect(parseRunnerProvider("RUNNER_PROVIDER_UNSPECIFIED")).toBeUndefined()
        expect(parseRunnerProvider("bitbucket")).toBeUndefined()
        expect(parseRunnerProvider(undefined)).toBeUndefined()

        expect(parseRunnerScope("REPO")).toBe(RunnerScope.REPO)
        expect(parseRunnerScope("ORG")).toBe(RunnerScope.ORG)
        expect(parseRunnerScope("RUNNER_SCOPE_UNSPECIFIED")).toBeUndefined()
        expect(parseRunnerScope("")).toBeUndefined()
    })

    it("prefills the form from the suggested defaults", () => {
        const form = initialRunnerForm({
            suggestedRunnerDefaults: {
                provider: "GITLAB",
                scope: "ORG",
                target: "acme",
                baseUrl: "https://gitlab.example",
                labels: ["a", "b"],
            },
        })

        expect(form).toMatchObject({
            provider: RunnerProvider.GITLAB,
            scope: RunnerScope.ORG,
            target: "acme",
            baseUrl: "https://gitlab.example",
            labels: "a, b",
            accessToken: "",
            registrationToken: "",
        })
    })

    it("falls back to the suggested pattern, then the image, for the provider", () => {
        const byPattern = initialRunnerForm({suggestedPattern: ServicePattern.SERVICE_PATTERN_GITLAB_RUNNER})
        const byImage = initialRunnerForm({imageName: "myoung34/github-runner:latest"})

        expect(byPattern.provider).toBe(RunnerProvider.GITLAB)
        expect(byImage.provider).toBe(RunnerProvider.GITHUB)
    })

    it("leaves the provider unset and the scope on repo when nothing is known", () => {
        const form = initialRunnerForm({imageName: "alpine"})

        expect(form.provider).toBeUndefined()
        expect(form.scope).toBe(RunnerScope.REPO)
        expect(form.target).toBe("")
    })
})
