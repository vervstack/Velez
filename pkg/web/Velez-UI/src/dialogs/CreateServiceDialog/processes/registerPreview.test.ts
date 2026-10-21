import {describe, expect, it} from "vitest"

import {ValueOrigin, type ContainerPreview} from "@/app/api/velez"
import {
    isAdded,
    previewOfContainer,
    previewSectionsOf,
} from "@/dialogs/CreateServiceDialog/processes/registerPreview.ts"

const PREVIEW: ContainerPreview = {
    containerId: "a",
    networkMode: {key: "mode", value: "bridge", origin: ValueOrigin.TRANSFERRED},
    env: [
        {key: "A", value: "1", origin: ValueOrigin.TRANSFERRED},
        {key: "VERV_SERVICE", value: "app", origin: ValueOrigin.ADDED},
    ],
    labels: [],
}

describe("registerPreview", () => {
    it("treats only ADDED values as added", () => {
        expect(isAdded({origin: ValueOrigin.ADDED})).toBe(true)
        expect(isAdded({origin: ValueOrigin.TRANSFERRED})).toBe(false)
        expect(isAdded({})).toBe(false)
    })

    it("finds the entry of a container by id", () => {
        const response = {containers: [{containerId: "b"}, PREVIEW]}

        expect(previewOfContainer(response, "a")).toBe(PREVIEW)
        expect(previewOfContainer(response, "zzz")).toBeUndefined()
        expect(previewOfContainer(undefined, "a")).toBeUndefined()
    })

    it("drops empty sections and keeps the section order", () => {
        const sections = previewSectionsOf(PREVIEW)

        expect(sections.map((section) => section.key)).toEqual(["networkMode", "env"])
    })

    it("lists added rows before transferred ones without reordering within a group", () => {
        const env = previewSectionsOf(PREVIEW).find((section) => section.key === "env")

        expect(env?.values.map((value) => value.key)).toEqual(["VERV_SERVICE", "A"])
    })

    it("returns no sections for an empty preview", () => {
        expect(previewSectionsOf({containerId: "a"})).toEqual([])
    })
})
