import {describe, expect, it} from "vitest"

import {RunnerProvider, RunnerScope} from "@/app/api/velez"

import {
    buildCreateRunnerRequest,
} from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/processes/buildCreateRunnerRequest.ts"

function newForm(overrides: Partial<Parameters<typeof buildCreateRunnerRequest>[0]> = {}) {
    return {
        name: "ci",
        provider: RunnerProvider.GITHUB,
        scope: RunnerScope.REPO,
        target: "org/repo",
        labels: "",
        environment: "",
        accessToken: "token",
        baseUrl: "",
        dockerImage: "",
        dindName: "ci-dind",
        dockerSocketAddress: "",
        concurrent: "",
        isBuildkitEnabled: false,
        ...overrides,
    }
}

describe("buildCreateRunnerRequest", () => {
    it("sends only dindName when a dind is chosen", () => {
        const req = buildCreateRunnerRequest(newForm())

        expect(req?.dindName).toBe("ci-dind")
        expect(req?.dockerSocketAddress).toBeUndefined()
    })

    it("sends only dockerSocketAddress when an external daemon is chosen", () => {
        const req = buildCreateRunnerRequest(newForm({dindName: "", dockerSocketAddress: "tcp://host:2375"}))

        expect(req?.dindName).toBeUndefined()
        expect(req?.dockerSocketAddress).toBe("tcp://host:2375")
    })

    it("returns null when no docker daemon is chosen", () => {
        expect(buildCreateRunnerRequest(newForm({dindName: ""}))).toBeNull()
    })

    it("returns null when both docker daemon fields are set", () => {
        expect(buildCreateRunnerRequest(newForm({dockerSocketAddress: "tcp://host:2375"}))).toBeNull()
    })

    it("forwards isBuildkitEnabled to the request", () => {
        const req = buildCreateRunnerRequest(newForm({isBuildkitEnabled: true}))

        expect(req?.isBuildkitEnabled).toBe(true)
    })
})
