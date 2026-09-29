import {describe, expect, it} from "vitest"

import {RunnerProvider, ServicePattern} from "@/app/api/velez"
import {
    adoptTitle,
    PRODUCT_CARDS,
    runnerProviderOf,
    screenTitle,
    suggestedScreenOf,
} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

describe("serviceScreen", () => {
    it("maps runner screens to their provider and every other screen to none", () => {
        expect(runnerProviderOf("githubRunner")).toBe(RunnerProvider.GITHUB)
        expect(runnerProviderOf("gitlabRunner")).toBe(RunnerProvider.GITLAB)
        expect(runnerProviderOf("postgres")).toBeUndefined()
        expect(runnerProviderOf("picker")).toBeUndefined()
    })

    it("titles the picker and every product screen", () => {
        expect(screenTitle("picker")).toBe("Create service")
        PRODUCT_CARDS.forEach((card) => expect(screenTitle(card.screen)).not.toBe(""))
    })

    it("maps a container's suggested pattern to the matching screen", () => {
        expect(suggestedScreenOf(ServicePattern.SERVICE_PATTERN_POSTGRES)).toBe("postgres")
        expect(suggestedScreenOf(ServicePattern.SERVICE_PATTERN_GITHUB_RUNNER)).toBe("githubRunner")
        expect(suggestedScreenOf(ServicePattern.SERVICE_PATTERN_GITLAB_RUNNER)).toBe("gitlabRunner")
        expect(suggestedScreenOf(ServicePattern.SERVICE_PATTERN_REGISTRY)).toBe("registry")
    })

    it("suggests nothing for an unspecified or missing pattern", () => {
        expect(suggestedScreenOf(ServicePattern.SERVICE_PATTERN_UNSPECIFIED)).toBeUndefined()
        expect(suggestedScreenOf(undefined)).toBeUndefined()
    })

    it("titles the adopt dialog with the container name", () => {
        expect(adoptTitle("db-1")).toBe("Register container db-1")
        expect(adoptTitle()).toBe("Register container")
    })
})
