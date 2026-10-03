import {describe, expect, it} from "vitest"

import {
    currentTagOf,
    imageTagToSend,
    isVersionTag,
    pickVersionTag,
    repositoryOf,
    sortVersionTags,
} from "@/dialogs/CreateServiceDialog/processes/imageVersion.ts"

describe("repositoryOf", () => {
    it("drops the tag", () => {
        expect(repositoryOf("nginx:latest")).toBe("nginx")
        expect(repositoryOf("ghcr.io/o/r:1")).toBe("ghcr.io/o/r")
    })

    it("keeps a registry port", () => {
        expect(repositoryOf("localhost:5000/x:1")).toBe("localhost:5000/x")
        expect(repositoryOf("localhost:5000/x")).toBe("localhost:5000/x")
    })

    it("drops a digest", () => {
        expect(repositoryOf("x@sha256:abc")).toBe("x")
        expect(repositoryOf("x:1@sha256:abc")).toBe("x")
    })

    it("returns an untagged name as is", () => {
        expect(repositoryOf("nginx")).toBe("nginx")
    })
})

describe("currentTagOf", () => {
    it("returns the tag part", () => {
        expect(currentTagOf("postgres:16")).toBe("16")
        expect(currentTagOf("localhost:5000/x:1.2")).toBe("1.2")
    })

    it("falls back to latest without a tag", () => {
        expect(currentTagOf("nginx")).toBe("latest")
        expect(currentTagOf("localhost:5000/x")).toBe("latest")
        expect(currentTagOf("x@sha256:abc")).toBe("latest")
    })
})

describe("isVersionTag", () => {
    it("accepts version-like tags", () => {
        for (const tag of ["1", "1.27", "1.27.2", "v1.2.3", "1.27.2-alpine"]) {
            expect(isVersionTag(tag)).toBe(true)
        }
    })

    it("rejects named tags", () => {
        for (const tag of ["latest", "stable", "main", "alpine"]) {
            expect(isVersionTag(tag)).toBe(false)
        }
    })
})

describe("sortVersionTags", () => {
    it("puts the most specific versions first and named tags last", () => {
        const sorted = sortVersionTags(["latest", "1", "1.27", "stable", "1.27.2"])

        expect(sorted).toEqual(["1.27.2", "1.27", "1", "latest", "stable"])
    })

    it("puts the longer tag first on equal specificity", () => {
        expect(sortVersionTags(["1.27.2", "1.27.2-alpine"])).toEqual(["1.27.2-alpine", "1.27.2"])
    })

    it("does not mutate its input", () => {
        const tags = ["1", "1.2"]
        sortVersionTags(tags)

        expect(tags).toEqual(["1", "1.2"])
    })
})

describe("pickVersionTag", () => {
    it("picks the most specific version", () => {
        expect(pickVersionTag(["latest", "1.27", "1.27.2"], "latest")).toBe("1.27.2")
    })

    it("falls back to the current tag when nothing looks like a version", () => {
        expect(pickVersionTag(["latest", "stable"], "latest")).toBe("latest")
        expect(pickVersionTag([], "latest")).toBe("latest")
    })
})

describe("imageTagToSend", () => {
    it("is undefined when the selection is the current tag", () => {
        expect(imageTagToSend("latest", "latest")).toBeUndefined()
    })

    it("is the selected tag otherwise", () => {
        expect(imageTagToSend("1.27.2", "latest")).toBe("1.27.2")
    })
})
