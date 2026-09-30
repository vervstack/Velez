import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {RunnerProvider, type DockerContainer} from "@/app/api/velez"
import {useGetContainerQuery} from "@/processes/queries/containers.ts"
import type {RunnerForm} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"
import RunnerAdoptScreen from "@/dialogs/CreateServiceDialog/screens/RunnerAdoptScreen/RunnerAdoptScreen.tsx"

interface MockProps {
    initialRunner: RunnerForm
    isRegistrationTokenFound: boolean
}

vi.mock("@/processes/queries/containers.ts", () => ({useGetContainerQuery: vi.fn()}))
vi.mock("@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx", () => ({
    default: ({initialRunner, isRegistrationTokenFound}: MockProps) => (
        <span>{`${initialRunner.provider} ${initialRunner.target} ${isRegistrationTokenFound}`}</span>
    ),
}))

type QueryResult = ReturnType<typeof useGetContainerQuery>

function renderScreen(query: Partial<QueryResult>, initialProvider?: RunnerProvider) {
    vi.mocked(useGetContainerQuery).mockReturnValue(query as Partial<QueryResult> as QueryResult)

    const container: DockerContainer = {id: "c1", name: "runner"}
    render(<RunnerAdoptScreen container={container} initialProvider={initialProvider} onBusyChange={vi.fn()}/>)
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("RunnerAdoptScreen", () => {
    it("offers a retry when the container fails to load", () => {
        const refetch = vi.fn()
        renderScreen({isError: true, refetch} as Partial<QueryResult>)

        fireEvent.click(screen.getByText("Retry"))

        expect(screen.getByText("Failed to load the container.")).toBeInTheDocument()
        expect(refetch).toHaveBeenCalledTimes(1)
    })

    it("prefills the form from the container's suggested defaults", () => {
        renderScreen({
            data: {
                suggestedRunnerDefaults: {provider: "GITLAB", target: "acme", isRegistrationTokenFound: true},
            },
        })

        expect(screen.getByText("GITLAB acme true")).toBeInTheDocument()
    })

    it("falls back to the picked provider when the container does not tell it", () => {
        renderScreen({data: {}}, RunnerProvider.GITHUB)

        expect(screen.getByText("GITHUB false")).toBeInTheDocument()
    })
})
