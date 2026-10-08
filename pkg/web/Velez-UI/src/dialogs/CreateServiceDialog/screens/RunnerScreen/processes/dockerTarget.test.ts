import {describe, expect, it} from "vitest"

import {
    EXTERNAL_DOCKER_CHOICE,
    isBuildkitSupported,
    resolveDockerTarget,
} from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/processes/dockerTarget.ts"

describe("resolveDockerTarget", () => {
    it("selects the dind and drops the external address when a dind is chosen", () => {
        expect(resolveDockerTarget("ci", "tcp://host:2375")).toEqual({dindName: "ci", dockerSocketAddress: ""})
    })

    it("selects the trimmed external address when the external choice is made", () => {
        expect(resolveDockerTarget(EXTERNAL_DOCKER_CHOICE, " tcp://host:2375 ")).toEqual({
            dindName: "",
            dockerSocketAddress: "tcp://host:2375",
        })
    })

    it("selects nothing when no choice is made", () => {
        expect(resolveDockerTarget("", "tcp://host:2375")).toEqual({dindName: "", dockerSocketAddress: ""})
    })
})

describe("isBuildkitSupported", () => {
    it("is unsupported for a dind and for an external daemon", () => {
        expect(isBuildkitSupported("ci")).toBe(false)
        expect(isBuildkitSupported(EXTERNAL_DOCKER_CHOICE)).toBe(false)
    })

    it("is supported when no daemon is chosen", () => {
        expect(isBuildkitSupported("")).toBe(true)
    })
})
