import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateRunnerMutation} from "@/processes/queries/runners.ts"
import {useListDindsQuery} from "@/processes/queries/dinds.ts"
import Button from "@/components/base/Button.tsx"
import RunnerScreen from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/RunnerScreen.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/runners.ts", () => ({
    CreateRunnerMutation: vi.fn(),
    RUNNERS_QUERY_KEY: ["runners"],
}))
vi.mock("@/processes/queries/dinds.ts", () => ({useListDindsQuery: vi.fn()}))
vi.mock("@/widgets/CreateDindForm/CreateDindForm.tsx", () => ({
    default: ({onCreated}: { onCreated(name: string): void }) => (
        <Button variant="primary" onClick={() => onCreated("fresh-daemon")}>create daemon form</Button>
    ),
}))
vi.mock("@/widgets/TaskProgressScreen/TaskProgressScreen.tsx", () => ({
    default: ({start}: { start(): Promise<unknown> }) => (
        <Button variant="primary" onClick={() => start()}>start task</Button>
    ),
}))

type RunnerMutation = ReturnType<typeof CreateRunnerMutation>
type DindsQuery = ReturnType<typeof useListDindsQuery>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function inputFor(label: string) {
    return screen.getByText(label).previousElementSibling as HTMLInputElement
}

const EXTERNAL_LABEL = "External Docker daemon (advanced)"
const ADDRESS_LABEL = "Docker daemon address (tcp://host:port)"

function renderScreen() {
    const mutateAsync = vi.fn().mockResolvedValue({})
    vi.mocked(CreateRunnerMutation).mockReturnValue({mutateAsync} as Partial<RunnerMutation> as RunnerMutation)
    vi.mocked(useListDindsQuery).mockReturnValue(
        {data: {dinds: [{name: "ci"}]}} as Partial<DindsQuery> as DindsQuery,
    )
    vi.mocked(useDialog).mockReturnValue({CloseDialog: vi.fn()} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)

    render(<RunnerScreen onBusyChange={vi.fn()}/>)
    return {mutateAsync}
}

function fillRequiredFields() {
    fireEvent.change(inputFor("Name"), {target: {value: "build-runner"}})
    fireEvent.change(inputFor("Target"), {target: {value: "org/repo"}})
    fireEvent.change(inputFor("GitHub Access Token"), {target: {value: "ghp_token"}})
}

function openDaemonDropdown() {
    fireEvent.click(screen.getAllByRole("button", {expanded: false}).at(-1) as HTMLElement)
}

function chooseDaemon(optionName: string) {
    openDaemonDropdown()
    fireEvent.mouseDown(screen.getByText(optionName))
}

function createButton() {
    return screen.getByRole("button", {name: "Create"})
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("RunnerScreen docker daemon choice", () => {
    it("keeps Create disabled until a Docker daemon is chosen", () => {
        renderScreen()
        fillRequiredFields()

        expect(createButton()).toBeDisabled()

        chooseDaemon("ci")

        expect(createButton()).toBeEnabled()
    })

    it("sends only dindName when an existing daemon is chosen", async () => {
        const {mutateAsync} = renderScreen()
        fillRequiredFields()
        chooseDaemon("ci")

        fireEvent.click(createButton())
        fireEvent.click(screen.getByText("start task"))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledTimes(1))
        const req = mutateAsync.mock.calls[0][0]
        expect(req.dindName).toBe("ci")
        expect(req.dockerSocketAddress).toBeUndefined()
    })

    it("keeps Create disabled for the external choice until an address is typed", () => {
        renderScreen()
        fillRequiredFields()

        chooseDaemon(EXTERNAL_LABEL)
        expect(createButton()).toBeDisabled()

        fireEvent.change(inputFor(ADDRESS_LABEL), {target: {value: "  "}})
        expect(createButton()).toBeDisabled()

        fireEvent.change(inputFor(ADDRESS_LABEL), {target: {value: "tcp://host:2375"}})
        expect(createButton()).toBeEnabled()
    })

    it("sends only dockerSocketAddress when the external daemon is chosen", async () => {
        const {mutateAsync} = renderScreen()
        fillRequiredFields()
        chooseDaemon(EXTERNAL_LABEL)
        fireEvent.change(inputFor(ADDRESS_LABEL), {target: {value: " tcp://host:2375 "}})

        fireEvent.click(createButton())
        fireEvent.click(screen.getByText("start task"))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledTimes(1))
        const req = mutateAsync.mock.calls[0][0]
        expect(req.dockerSocketAddress).toBe("tcp://host:2375")
        expect(req.dindName).toBeUndefined()
    })

    it("drops a typed external address when the user switches back to a named daemon", async () => {
        const {mutateAsync} = renderScreen()
        fillRequiredFields()
        chooseDaemon(EXTERNAL_LABEL)
        fireEvent.change(inputFor(ADDRESS_LABEL), {target: {value: "tcp://host:2375"}})

        chooseDaemon("ci")
        fireEvent.click(createButton())
        fireEvent.click(screen.getByText("start task"))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledTimes(1))
        const req = mutateAsync.mock.calls[0][0]
        expect(req.dindName).toBe("ci")
        expect(req.dockerSocketAddress).toBeUndefined()
    })

    it("selects the freshly created daemon after the inline create form finishes", async () => {
        const {mutateAsync} = renderScreen()
        fillRequiredFields()
        openDaemonDropdown()
        fireEvent.mouseDown(screen.getByText("Create new Docker daemon"))

        fireEvent.click(screen.getByText("create daemon form"))
        fireEvent.click(createButton())
        fireEvent.click(screen.getByText("start task"))

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledTimes(1))
        expect(mutateAsync.mock.calls[0][0].dindName).toBe("fresh-daemon")
    })
})
