import {describe, expect, it} from "vitest"

import {RunnerProvider} from "@/app/api/velez"
import {PRODUCT_CARDS, runnerProviderOf, screenTitle} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

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
})
