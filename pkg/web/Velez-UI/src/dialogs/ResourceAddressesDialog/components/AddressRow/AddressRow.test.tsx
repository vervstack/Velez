import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"
import {useQuery} from "@tanstack/react-query"

import type {ResourceAddress} from "@/model/service_page/ServicePageModel"
import AddressRow from "@/dialogs/ResourceAddressesDialog/components/AddressRow/AddressRow.tsx"

vi.mock("@tanstack/react-query", () => ({useQuery: vi.fn()}))

type PingQuery = ReturnType<typeof useQuery>

function renderRow(address: ResourceAddress, query: Partial<PingQuery>) {
    vi.mocked(useQuery).mockReturnValue(query as Partial<PingQuery> as PingQuery)
    render(<AddressRow address={address}/>)
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("AddressRow", () => {
    it("links to the address host and port in a new tab", () => {
        renderRow({host: "10.0.0.5", port: 3909, scope: "docker"}, {isLoading: false, data: true})

        const link = screen.getByRole("link", {name: "http://10.0.0.5:3909"})
        expect(link).toHaveAttribute("href", "http://10.0.0.5:3909")
        expect(link).toHaveAttribute("target", "_blank")
    })

    it("falls back to the page hostname when the host is empty", () => {
        renderRow({host: "", port: 3909, scope: "docker"}, {isLoading: false, data: true})

        expect(screen.getByRole("link")).toHaveAttribute("href", `http://${window.location.hostname}:3909`)
    })

    it("labels docker addresses Public and vcn addresses VCN", () => {
        renderRow({host: "h", port: 1, scope: "docker"}, {isLoading: false, data: true})
        expect(screen.getByText("Public")).toBeInTheDocument()
    })

    it("labels vcn addresses VCN", () => {
        renderRow({host: "h", port: 1, scope: "vcn"}, {isLoading: false, data: true})
        expect(screen.getByText("VCN")).toBeInTheDocument()
    })

    it("shows a checking dot while the ping is loading", () => {
        renderRow({host: "h", port: 1, scope: "docker"}, {isLoading: true})
        expect(screen.getByRole("img", {name: "Checking"})).toBeInTheDocument()
    })

    it("shows a reachable dot when the ping resolves true", () => {
        renderRow({host: "h", port: 1, scope: "docker"}, {isLoading: false, data: true})
        expect(screen.getByRole("img", {name: "Reachable"})).toBeInTheDocument()
    })

    it("shows an unreachable dot when the ping resolves false", () => {
        renderRow({host: "h", port: 1, scope: "docker"}, {isLoading: false, data: false})
        expect(screen.getByRole("img", {name: "Unreachable"})).toBeInTheDocument()
    })
})
