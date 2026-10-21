import {describe, expect, it} from "vitest"

import {buildPreviewRequest, PreviewForm} from "@/dialogs/CreateServiceDialog/processes/buildPreviewRequest.ts"
import {initialRoles, pickRoot, pickService} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

function newForm(overrides: Partial<PreviewForm> = {}): PreviewForm {
    return {
        containerId: "a",
        environment: "",
        serviceName: " app ",
        pattern: "generic",
        roles: initialRoles("a"),
        ...overrides,
    }
}

describe("buildPreviewRequest", () => {
    it("builds a generic request with a trimmed name and no role ids by default", () => {
        expect(buildPreviewRequest(newForm())).toEqual({
            containerId: "a",
            environment: undefined,
            serviceName: "app",
            generic: {},
        })
    })

    it("returns nothing while the service name is empty", () => {
        expect(buildPreviewRequest(newForm({serviceName: "  "}))).toBeNull()
    })

    it("carries the chosen pattern kind", () => {
        expect(buildPreviewRequest(newForm({pattern: "postgres"}))).toHaveProperty("pg", {})
        expect(buildPreviewRequest(newForm({pattern: "registry"}))).toHaveProperty("registry", {})
        expect(buildPreviewRequest(newForm({pattern: "runner"}))).toHaveProperty("runner", {})
    })

    it("sends the service container id and forces generic when another member is the service", () => {
        const req = buildPreviewRequest(newForm({pattern: "postgres", roles: pickService(initialRoles("a"), "b")}))

        expect(req).toMatchObject({containerId: "a", serviceContainerId: "b", generic: {}})
        expect(req).not.toHaveProperty("pg")
        expect(req).not.toHaveProperty("rootContainerId")
    })

    it("sends only the root id when the service follows the switched root", () => {
        const req = buildPreviewRequest(newForm({roles: pickRoot(initialRoles("a"), "b", ["a", "b"])}))

        expect(req).toMatchObject({containerId: "a", rootContainerId: "b", generic: {}})
        expect(req).not.toHaveProperty("serviceContainerId")
    })
})
