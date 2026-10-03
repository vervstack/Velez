import {describe, expect, it} from "vitest"

import type {DockerContainer} from "@/app/api/velez"
import type {LayoutEntry, LayoutItem} from "@/pages/services/processes/groupContainers.ts"
import {containersOfEntry, itemsOfEntry, partitionByRegistration} from "@/pages/services/processes/partitionByRegistration.ts"

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

    it("keeps a fully registered compose group together", () => {
        const group: LayoutEntry = {kind: "compose", project: "p", items: [single("a", true), single("b", true)]}

        const result = partitionByRegistration([group], "all")

        expect(result.registered).toEqual([group])
        expect(result.unregistered).toEqual([])
    })

    it("splits a mixed compose group into registered and unregistered compose groups", () => {
        const group: LayoutEntry = {
            kind: "compose",
            project: "q",
            items: [single("a", true), single("b", false), single("c", true), single("d", false)],
        }

        const result = partitionByRegistration([group], "all")

        expect(result.registered).toEqual([
            {kind: "compose", project: "q", items: [single("a", true), single("c", true)]},
        ])
        expect(result.unregistered).toEqual([
            {kind: "compose", project: "q", items: [single("b", false), single("d", false)]},
        ])
    })

    it("emits a lone remaining item as a standalone entry instead of a compose wrapper", () => {
        const group: LayoutEntry = {
            kind: "compose",
            project: "q",
            items: [single("a", true), single("b", true), single("c", false)],
        }

        const result = partitionByRegistration([group], "all")

        expect(result.registered).toEqual([
            {kind: "compose", project: "q", items: [single("a", true), single("b", true)]},
        ])
        expect(result.unregistered).toEqual([single("c", false)])
    })
})

describe("partitionByRegistration awaiting onboarding end", () => {
    const leftover: LayoutItem = {kind: "single", container: {id: "old", isRegistered: false, replacedByContainerId: "new"}}
    const entries: LayoutEntry[] = [single("a", true), single("b", false), leftover]

    it("puts a leftover only into the awaiting-end list when the filter is all", () => {
        const result = partitionByRegistration(entries, "all")

        expect(result.awaitingEnd).toEqual([leftover])
        expect(result.registered).toEqual([entries[0]])
        expect(result.unregistered).toEqual([entries[1]])
    })

    it("keeps the awaiting-end list for the unregistered filter", () => {
        expect(partitionByRegistration(entries, "unregistered").awaitingEnd).toEqual([leftover])
    })

    it("hides the awaiting-end list for the registered filter", () => {
        expect(partitionByRegistration(entries, "registered").awaitingEnd).toEqual([])
    })

    it("treats a group containing a leftover as awaiting end even if the rest is registered", () => {
        const group = network({id: "root", isRegistered: true}, [{id: "s1", isRegistered: true, replacedByContainerId: "n"}])

        const result = partitionByRegistration([group], "all")

        expect(result.awaitingEnd).toEqual([group])
        expect(result.registered).toEqual([])
    })
})

describe("partitionByRegistration compose splitting", () => {
    const leftoverRoot: DockerContainer = {id: "root", isRegistered: true}
    const leftoverSidecar: DockerContainer = {id: "side", isRegistered: true, replacedByContainerId: "new"}

    it("keeps never-onboarded containers of a compose project out of the awaiting-end group", () => {
        const leftoverNetwork = network(leftoverRoot, [leftoverSidecar])
        const group: LayoutEntry = {
            kind: "compose",
            project: "p",
            items: [leftoverNetwork, single("x", false), single("y", false)],
        }

        const result = partitionByRegistration([group], "all")

        expect(result.awaitingEnd).toEqual([{kind: "compose", project: "p", items: [leftoverNetwork]}])
        expect(result.awaitingEnd.flatMap(containersOfEntry)).toHaveLength(2)
        expect(result.unregistered).toEqual([{kind: "compose", project: "p", items: [single("x", false), single("y", false)]}])
        expect(result.registered).toEqual([])
    })

    it("keeps a fully leftover compose project as one compose group", () => {
        const group: LayoutEntry = {
            kind: "compose",
            project: "p",
            items: [
                {kind: "single", container: {id: "a", replacedByContainerId: "n1"}},
                {kind: "single", container: {id: "b", replacedByContainerId: "n2"}},
            ],
        }

        const result = partitionByRegistration([group], "all")

        expect(result.awaitingEnd).toEqual([group])
        expect(result.unregistered).toEqual([])
    })

    it("preserves the original order of items inside each bucket", () => {
        const group: LayoutEntry = {
            kind: "compose",
            project: "p",
            items: [single("u1", false), single("r1", true), single("u2", false), single("r2", true), single("u3", false)],
        }

        const result = partitionByRegistration([group], "all")

        const unregisteredIds = result.unregistered.flatMap(itemsOfEntry).flatMap(containersOfEntry).map(toId)
        const registeredIds = result.registered.flatMap(itemsOfEntry).flatMap(containersOfEntry).map(toId)
        expect(unregisteredIds).toEqual(["u1", "u2", "u3"])
        expect(registeredIds).toEqual(["r1", "r2"])
    })

    it("applies the filter to the split buckets", () => {
        const group: LayoutEntry = {
            kind: "compose",
            project: "p",
            items: [network(leftoverRoot, [leftoverSidecar]), single("x", false), single("y", true), single("z", true)],
        }

        const onlyRegistered = partitionByRegistration([group], "registered")

        expect(onlyRegistered.awaitingEnd).toEqual([])
        expect(onlyRegistered.unregistered).toEqual([])
        expect(onlyRegistered.registered).toHaveLength(1)
    })
})

function toId(container: DockerContainer): string | undefined {
    return container.id
}

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
