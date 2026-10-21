import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import {useImageVersionsQuery} from "@/processes/queries/containers.ts"
import ImageVersionPicker from "@/dialogs/CreateServiceDialog/components/ImageVersionPicker/ImageVersionPicker.tsx"

vi.mock("@/processes/queries/containers.ts", () => ({useImageVersionsQuery: vi.fn()}))

type QueryResult = ReturnType<typeof useImageVersionsQuery>

interface RenderOptions {
    query: Partial<QueryResult>
    image?: string
    value?: string
}

function renderPicker({query, image = "nginx:latest", value}: RenderOptions) {
    vi.mocked(useImageVersionsQuery).mockReturnValue(query as Partial<QueryResult> as QueryResult)

    const onChange = vi.fn()
    render(<ImageVersionPicker containerId="c1" image={image} value={value} onChange={onChange}/>)
    return {onChange}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ImageVersionPicker", () => {
    it("shows the container's image and tag as read-only values while loading", () => {
        const {onChange} = renderPicker({query: {isLoading: true}, image: "nginx:1.27"})

        expect(screen.getByDisplayValue("nginx")).toBeDisabled()
        expect(screen.getByDisplayValue("1.27")).toBeDisabled()
        expect(onChange).not.toHaveBeenCalled()
    })

    it("shows a skeleton instead of the fields while the image is unknown", () => {
        renderPicker({query: {isLoading: true}, image: ""})

        expect(screen.queryByText("Image")).not.toBeInTheDocument()
        expect(screen.queryByText("Version")).not.toBeInTheDocument()
    })

    it("selects and reports the most specific version by default", () => {
        const {onChange} = renderPicker({
            query: {isLoading: false, isError: false, data: {tags: ["latest", "1.27", "1.27.2"]}},
        })

        expect(screen.getByDisplayValue("nginx")).toBeInTheDocument()
        expect(screen.getByText("Version")).toBeInTheDocument()
        expect(screen.getByText("1.27.2")).toBeInTheDocument()
        expect(onChange).toHaveBeenCalledWith("1.27.2")
    })

    it("falls back to the read-only current tag on error", () => {
        const {onChange} = renderPicker({query: {isLoading: false, isError: true}})

        expect(screen.getByDisplayValue("nginx")).toBeInTheDocument()
        expect(screen.getByDisplayValue("latest")).toBeDisabled()
        expect(onChange).not.toHaveBeenCalled()
    })

    it("falls back to the read-only current tag when no tags resolve", () => {
        renderPicker({query: {isLoading: false, isError: false, data: {tags: []}}})

        expect(screen.getByDisplayValue("nginx")).toBeInTheDocument()
        expect(screen.getByDisplayValue("latest")).toBeDisabled()
    })

    it("hints about pinning when the selection differs from the current tag", () => {
        renderPicker({query: {isLoading: false, isError: false, data: {tags: ["latest", "1.27.2"]}}})

        expect(screen.getByText("Pinned to this version when onboarded. The running image is the same."))
            .toBeInTheDocument()
    })

    it("shows no hint when the selection is the current tag", () => {
        renderPicker({
            query: {isLoading: false, isError: false, data: {tags: ["latest", "1.27.2"]}},
            value: "latest",
        })

        expect(screen.queryByText(/Pinned to this version/)).not.toBeInTheDocument()
    })
})
