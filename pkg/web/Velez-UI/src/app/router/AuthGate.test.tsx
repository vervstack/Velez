import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen, waitFor} from "@testing-library/react"
import {QueryClient, QueryClientProvider} from "@tanstack/react-query"
import {createMemoryRouter, RouterProvider} from "react-router-dom"
import {createElement} from "react"

import AuthGate from "@/app/router/AuthGate.tsx"

const ping = vi.fn()

vi.mock("@/processes/api/velez.ts", () => ({
    velezService: {ping: () => ping()},
}))

afterEach(() => {
    vi.clearAllMocks()
})

function renderGate() {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})
    const router = createMemoryRouter([
        {
            path: "/",
            element: createElement(AuthGate),
            children: [{index: true, element: createElement("div", null, "DASHBOARD")}],
        },
        {path: "/login", element: createElement("div", null, "LOGIN SCREEN")},
    ], {initialEntries: ["/"]})

    return render(createElement(QueryClientProvider, {client: queryClient}, createElement(RouterProvider, {router})))
}

describe("AuthGate", () => {
    it("redirects to the login screen when the version ping is rejected", async () => {
        ping.mockRejectedValue(new Error("permission denied"))

        renderGate()

        await waitFor(() => expect(screen.getByText("LOGIN SCREEN")).toBeInTheDocument())
        expect(screen.queryByText("DASHBOARD")).not.toBeInTheDocument()
    })

    it("renders the protected content when the version ping resolves", async () => {
        ping.mockResolvedValue({version: "1.2.3"})

        renderGate()

        await waitFor(() => expect(screen.getByText("DASHBOARD")).toBeInTheDocument())
        expect(screen.queryByText("LOGIN SCREEN")).not.toBeInTheDocument()
    })

    it("shows neither the protected content nor the login screen while the ping is in flight", async () => {
        function noop() {}

        let resolvePing: (value: unknown) => void = noop
        ping.mockReturnValue(new Promise((resolve) => {
            resolvePing = resolve
        }))

        renderGate()

        expect(screen.queryByText("DASHBOARD")).not.toBeInTheDocument()
        expect(screen.queryByText("LOGIN SCREEN")).not.toBeInTheDocument()

        resolvePing({version: "1.2.3"})
        await waitFor(() => expect(screen.getByText("DASHBOARD")).toBeInTheDocument())
    })
})
