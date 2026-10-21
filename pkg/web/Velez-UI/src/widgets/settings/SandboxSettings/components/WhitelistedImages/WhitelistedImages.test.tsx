import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useListWhitelistedImagesQuery} from "@/processes/queries/settings.ts"
import WhitelistedImages from "@/widgets/settings/SandboxSettings/components/WhitelistedImages/WhitelistedImages.tsx"

vi.mock("@/processes/queries/settings.ts", () => ({useListWhitelistedImagesQuery: vi.fn()}))
vi.mock(
    "@/widgets/settings/SandboxSettings/components/WhitelistedImages/components/WhitelistedImageRow/WhitelistedImageRow.tsx",
    () => ({
        default: ({image}: {image: {repository?: string}}) => <div>{image.repository}</div>,
    }),
)

type ImagesQuery = ReturnType<typeof useListWhitelistedImagesQuery>

function renderImages(query: Partial<ImagesQuery> = {}) {
    const refetch = vi.fn()
    vi.mocked(useListWhitelistedImagesQuery).mockReturnValue(
        {refetch, ...query} as Partial<ImagesQuery> as ImagesQuery,
    )

    render(<WhitelistedImages/>)
    return {refetch}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("WhitelistedImages", () => {
    it("renders the title and description", () => {
        renderImages({data: {images: []}})

        expect(screen.getByText("Whitelisted images")).toBeInTheDocument()
        expect(screen.getByText(/Images allowed to run with elevated access/)).toBeInTheDocument()
    })

    it("renders a row for every image returned by the query", () => {
        renderImages({data: {images: [
            {repository: "tailscale/tailscale", isBuiltin: true},
            {id: "1", repository: "my/image"},
        ]}})

        expect(screen.getByText("tailscale/tailscale")).toBeInTheDocument()
        expect(screen.getByText("my/image")).toBeInTheDocument()
    })

    it("shows no rows and no error while the list is loading", () => {
        renderImages({isLoading: true})

        expect(screen.queryByText("Retry")).not.toBeInTheDocument()
        expect(screen.queryByText("my/image")).not.toBeInTheDocument()
    })

    it("shows a retry-capable error state that refetches when the list fails to load", () => {
        const {refetch} = renderImages({isError: true})

        expect(screen.getByText("Failed to load whitelisted images.")).toBeInTheDocument()

        fireEvent.click(screen.getByText("Retry"))

        expect(refetch).toHaveBeenCalledTimes(1)
    })
})
