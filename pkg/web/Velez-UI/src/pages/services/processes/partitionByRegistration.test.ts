import {describe, expect, it} from "vitest"

import type {DockerContainer} from "@/app/api/velez"
import type {LayoutEntry, LayoutItem} from "@/pages/services/processes/groupContainers.ts"
import {containersOfEntry, partitionByRegistration} from "@/pages/services/processes/partitionByRegistration.ts"

function single(id: string, isRegistered: boolean): LayoutItem {
    return {kind: "single", container: {id, isRegistered}}
}

function network(root: DockerContainer, members: DockerContainer[]): LayoutItem {
    return {kind: "network", root, members}
}

describe("partitionByRegistration", () => {
    const entries: LayoutEntry[] = [single("a", true), single("b", false), single("c", true)]

    it("splits entries into registered and unregistered when the filter is all", () => {
        const result = partitionByRegistration(entries, "all")

        expect(result.registered).toEqual([entries[0], entries[2]])
        expect(result.unregistered).toEqual([entries[1]])
    })

    it("drops unregistered entries when the filter is registered", () => {
        const result = partitionByRegistration(entries, "registered")

        expect(result.registered).toHaveLength(2)
        expect(result.unregistered).toHaveLength(0)
    })

    it("drops registered entries when the filter is unregistered", () => {
        const result = partitionByRegistration(entries, "unregistered")

        expect(result.registered).toHaveLength(0)
        expect(result.unregistered).toEqual([entries[1]])
    })

    it("treats a network group with an unregistered member as unregistered", () => {
        const group = network({id: "root", isRegistered: true}, [{id: "s1", isRegistered: false}])

        const result = partitionByRegistration([group], "all")

        expect(result.unregistered).toEqual([group])
    })

    it("treats a compose group as registered only when every container inside is registered", () => {
        const registeredGroup: LayoutEntry = {kind: "compose", project: "p", items: [single("a", true)]}
        const mixedGroup: LayoutEntry = {kind: "compose", project: "q", items: [single("b", true), single("c", false)]}

        const result = partitionByRegistration([registeredGroup, mixedGroup], "all")

        expect(result.registered).toEqual([registeredGroup])
        expect(result.unregistered).toEqual([mixedGroup])
    })
})

describe("containersOfEntry", () => {
    it("flattens network members inside a compose group", () => {
        const entry: LayoutEntry = {
            kind: "compose",
            project: "p",
            items: [network({id: "root"}, [{id: "s1"}]), single("x", false)],
        }

        const ids = containersOfEntry(entry).map(function toId(container) {
            return container.id
        })

        expect(ids).toEqual(["root", "s1", "x"])
    })
})
