import {describe, expect, it} from "vitest"

import type {DockerContainer} from "@/app/api/velez"
import {groupContainers} from "@/pages/services/processes/groupContainers.ts"

const PROJECT_LABEL = "com.docker.compose.project"

function container(id: string, extra: Partial<DockerContainer> = {}): DockerContainer {
    return {id, name: id, ...extra}
}

function inProject(project: string): Partial<DockerContainer> {
    return {labels: {[PROJECT_LABEL]: project}}
}

describe("groupContainers", () => {
    it("returns plain singles in original order when nothing is grouped", () => {
        const result = groupContainers([container("a"), container("b"), container("c")])

        expect(result.map((e) => e.kind)).toEqual(["single", "single", "single"])
        expect(result).toEqual([
            {kind: "single", container: container("a")},
            {kind: "single", container: container("b")},
            {kind: "single", container: container("c")},
        ])
    })

    it("groups containers sharing a compose project", () => {
        const a = container("a", inProject("shop"))
        const b = container("b")
        const c = container("c", inProject("shop"))

        const result = groupContainers([a, b, c])

        expect(result).toEqual([
            {
                kind: "compose",
                project: "shop",
                items: [{kind: "single", container: a}, {kind: "single", container: c}],
            },
            {kind: "single", container: b},
        ])
    })

    it("leaves a lone compose container as a plain single", () => {
        const a = container("a", inProject("shop"))

        expect(groupContainers([a])).toEqual([{kind: "single", container: a}])
    })

    it("ignores labels other than the compose project", () => {
        const a = container("a", {labels: {"com.docker.compose.service": "web"}})
        const b = container("b", {labels: {"com.docker.compose.service": "web"}})

        expect(groupContainers([a, b]).map((e) => e.kind)).toEqual(["single", "single"])
    })

    it("builds a network group from a root and its sidecars", () => {
        const root = container("root")
        const side = container("side", {networkOwnerContainerId: "root"})
        const other = container("other")

        const result = groupContainers([side, other, root])

        expect(result).toEqual([
            {kind: "network", root, members: [side]},
            {kind: "single", container: other},
        ])
    })

    it("treats a container whose network owner is not listed as a single", () => {
        const side = container("side", {networkOwnerContainerId: "missing"})

        expect(groupContainers([side])).toEqual([{kind: "single", container: side}])
    })

    it("nests a network group inside the compose project of its root", () => {
        const root = container("root", inProject("shop"))
        const side = container("side", {networkOwnerContainerId: "root"})
        const web = container("web", inProject("shop"))

        const result = groupContainers([root, side, web])

        expect(result).toEqual([
            {
                kind: "compose",
                project: "shop",
                items: [
                    {kind: "network", root, members: [side]},
                    {kind: "single", container: web},
                ],
            },
        ])
    })

    it("makes a network group with a compose root a compose group on its own", () => {
        const root = container("root", inProject("shop"))
        const side = container("side", {networkOwnerContainerId: "root"})

        const result = groupContainers([root, side])

        expect(result).toEqual([
            {kind: "compose", project: "shop", items: [{kind: "network", root, members: [side]}]},
        ])
    })

    it("ignores the compose project of a sidecar when its root has none", () => {
        const root = container("root")
        const side = container("side", {networkOwnerContainerId: "root", ...inProject("shop")})

        expect(groupContainers([root, side])).toEqual([{kind: "network", root, members: [side]}])
    })

    it("resolves a chain of network owners to the ultimate root", () => {
        const root = container("root")
        const mid = container("mid", {networkOwnerContainerId: "root"})
        const leaf = container("leaf", {networkOwnerContainerId: "mid"})

        const result = groupContainers([leaf, mid, root])

        expect(result).toEqual([{kind: "network", root, members: [leaf, mid]}])
    })

    it("does not loop on circular network owners", () => {
        const a = container("a", {networkOwnerContainerId: "b"})
        const b = container("b", {networkOwnerContainerId: "a"})

        expect(groupContainers([a, b]).map((e) => e.kind)).toEqual(["single", "single"])
    })

    it("keeps the original list order across groups and singles", () => {
        const first = container("first")
        const shopA = container("shopA", inProject("shop"))
        const root = container("root")
        const side = container("side", {networkOwnerContainerId: "root"})
        const shopB = container("shopB", inProject("shop"))

        const result = groupContainers([first, shopA, root, side, shopB])

        expect(result.map((e) => e.kind)).toEqual(["single", "compose", "network"])
    })
})
