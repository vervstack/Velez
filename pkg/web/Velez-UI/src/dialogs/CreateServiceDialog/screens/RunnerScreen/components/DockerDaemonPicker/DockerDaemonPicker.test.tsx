import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useListDindsQuery} from "@/processes/queries/dinds.ts"
import DockerDaemonPicker
    from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/components/DockerDaemonPicker/DockerDaemonPicker.tsx"
import {
    EXTERNAL_DOCKER_CHOICE,
} from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/processes/dockerTarget.ts"

vi.mock("@/processes/queries/dinds.ts", () => ({useListDindsQuery: vi.fn()}))

type DindsQuery = ReturnType<typeof useListDindsQuery>

function inputFor(label: string) {
    return screen.getByText(label).previousElementSibling as HTMLInputElement
}

interface RenderOptions {
    query?: Partial<DindsQuery>
    choice?: string
    externalAddress?: string
}

const EXTERNAL_LABEL = "External Docker daemon (advanced)"
const ADDRESS_LABEL = "Docker daemon address (tcp://host:port)"

function renderPicker({query, choice = "", externalAddress = ""}: RenderOptions = {}) {
    const refetch = vi.fn()
    const onChoiceChange = vi.fn()
    const onExternalAddressChange = vi.fn()
    const onCreateNew = vi.fn()
    vi.mocked(useListDindsQuery).mockReturnValue({
        refetch,
        data: {dinds: [{name: "ci"}, {name: "staging"}]},
        ...query,
    } as Partial<DindsQuery> as DindsQuery)

    render(
        <DockerDaemonPicker
            choice={choice}
            externalAddress={externalAddress}
            isDisabled={false}
            onChoiceChange={onChoiceChange}
            onExternalAddressChange={onExternalAddressChange}
            onCreateNew={onCreateNew}
        />
    )
    return {refetch, onChoiceChange, onExternalAddressChange, onCreateNew}
}

function openDropdown() {
    fireEvent.click(screen.getByRole("button", {expanded: false}))
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("DockerDaemonPicker", () => {
    it("lists the existing daemons and the external option", () => {
        renderPicker()

        openDropdown()

        expect(screen.getByText("ci")).toBeInTheDocument()
        expect(screen.getByText("staging")).toBeInTheDocument()
        expect(screen.getByText(EXTERNAL_LABEL)).toBeInTheDocument()
    })

    it("reports the daemon name when an existing daemon is chosen", () => {
        const {onChoiceChange} = renderPicker()
        openDropdown()

        fireEvent.mouseDown(screen.getByText("staging"))

        expect(onChoiceChange).toHaveBeenCalledWith("staging")
    })

    it("reports the external choice when the external option is chosen", () => {
        const {onChoiceChange} = renderPicker()
        openDropdown()

        fireEvent.mouseDown(screen.getByText(EXTERNAL_LABEL))

        expect(onChoiceChange).toHaveBeenCalledWith(EXTERNAL_DOCKER_CHOICE)
    })

    it("hides the address input while a named daemon is chosen", () => {
        renderPicker({choice: "ci"})

        expect(screen.queryByText(ADDRESS_LABEL)).not.toBeInTheDocument()
    })

    it("reveals the address input and risk notice when the external option is chosen", () => {
        const {onExternalAddressChange} = renderPicker({choice: EXTERNAL_DOCKER_CHOICE})

        fireEvent.change(inputFor(ADDRESS_LABEL), {target: {value: "tcp://host:2375"}})

        expect(onExternalAddressChange).toHaveBeenCalledWith("tcp://host:2375")
        expect(screen.getByText(/Only use a daemon you trust/)).toBeInTheDocument()
    })

    it("calls onCreateNew from the footer action", () => {
        const {onCreateNew} = renderPicker()
        openDropdown()

        fireEvent.mouseDown(screen.getByText("Create new Docker daemon"))

        expect(onCreateNew).toHaveBeenCalledTimes(1)
    })

    it("shows nothing interactive while the daemons are loading", () => {
        renderPicker({query: {isLoading: true, data: undefined}})

        expect(screen.queryByRole("button")).not.toBeInTheDocument()
    })

    it("shows a retry-capable error state that refetches when the daemons fail to load", () => {
        const {refetch} = renderPicker({query: {isError: true, data: undefined}})

        expect(screen.getByText("Failed to load Docker daemons.")).toBeInTheDocument()
        fireEvent.click(screen.getByText("Retry"))

        expect(refetch).toHaveBeenCalledTimes(1)
    })
})
