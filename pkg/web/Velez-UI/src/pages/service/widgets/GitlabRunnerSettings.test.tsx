import {beforeEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, within} from "@testing-library/react"
import type {DropdownProps} from "@vervstack/chures"

import GitlabRunnerSettings from "@/pages/service/widgets/GitlabRunnerSettings.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {GetRunnerConfigResponse, RunnerLogLevel, RunnerPullPolicy} from "@/app/api/velez"
import {
    GetRunnerConfigQuery,
    GetRunnerCredentialsQuery,
    RedeployRunnerMutation,
    ReregisterRunnerMutation,
    UpdateRunnerConfigMutation,
} from "@/processes/queries/runners.ts"

vi.mock("@/processes/queries/runners.ts", () => ({
    GetRunnerConfigQuery: vi.fn(),
    GetRunnerCredentialsQuery: vi.fn(),
    UpdateRunnerConfigMutation: vi.fn(),
    ReregisterRunnerMutation: vi.fn(),
    RedeployRunnerMutation: vi.fn(),
}))

vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))

vi.mock("@vervstack/chures", async (importOriginal) => {
    const original = await importOriginal<typeof import("@vervstack/chures")>()

    function DropdownStub({label, options, value, onChange}: DropdownProps) {
        return (
            <div role="group" aria-label={label}>
                {(options ?? []).map((option) => {
                    if (typeof option !== "object" || !("id" in option)) return null

                    return (
                        <button
                            key={option.id}
                            aria-pressed={value.includes(option.id)}
                            onClick={() => onChange([option.id])}
                        >
                            {option.name}
                        </button>
                    )
                })}
            </div>
        )
    }

    return {...original, Dropdown: DropdownStub}
})

type ConfigQ = ReturnType<typeof GetRunnerConfigQuery>
type CredsQ = ReturnType<typeof GetRunnerCredentialsQuery>
type UpdateM = ReturnType<typeof UpdateRunnerConfigMutation>
type ReregM = ReturnType<typeof ReregisterRunnerMutation>
type RedeployM = ReturnType<typeof RedeployRunnerMutation>
type ToasterS = ReturnType<typeof useToaster>

const ALWAYS = RunnerPullPolicy.RUNNER_PULL_POLICY_ALWAYS
const IF_NOT_PRESENT = RunnerPullPolicy.RUNNER_PULL_POLICY_IF_NOT_PRESENT
const NEVER = RunnerPullPolicy.RUNNER_PULL_POLICY_NEVER

function renderSettings(config: GetRunnerConfigResponse) {
    const mutate = vi.fn()

    vi.mocked(GetRunnerConfigQuery).mockReturnValue(
        {data: config, isLoading: false, isError: false} as Partial<ConfigQ> as ConfigQ,
    )
    vi.mocked(GetRunnerCredentialsQuery).mockReturnValue(
        {data: undefined, isFetching: false} as Partial<CredsQ> as CredsQ,
    )
    vi.mocked(UpdateRunnerConfigMutation).mockReturnValue(
        {mutate, isPending: false} as Partial<UpdateM> as UpdateM,
    )
    vi.mocked(ReregisterRunnerMutation).mockReturnValue(
        {mutate: vi.fn(), isPending: false} as Partial<ReregM> as ReregM,
    )
    vi.mocked(RedeployRunnerMutation).mockReturnValue(
        {mutate: vi.fn(), isPending: false} as Partial<RedeployM> as RedeployM,
    )
    vi.mocked(useToaster).mockReturnValue(
        {bake: vi.fn(), catchGrpc: vi.fn()} as Partial<ToasterS> as ToasterS,
    )

    render(<GitlabRunnerSettings serviceName="runner-1"/>)
    return {mutate}
}

function pressedButtons(groupName: string): string[] {
    return within(screen.getByRole("group", {name: groupName}))
        .getAllByRole("button")
        .filter((button) => button.getAttribute("aria-pressed") === "true")
        .map((button) => button.textContent ?? "")
}

describe("GitlabRunnerSettings", () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    it("shows the loaded pull policy chain when the config has one", () => {
        renderSettings({pullPolicy: [IF_NOT_PRESENT, NEVER], logLevel: RunnerLogLevel.RUNNER_LOG_LEVEL_WARN})

        expect(pressedButtons("Policy")).toEqual(["If not present"])
        expect(pressedButtons("Fallback 1")).toEqual(["Never"])
        expect(pressedButtons("Level")).toEqual(["Warn"])
    })

    it("shows the default hint when the config has no pull policy", () => {
        renderSettings({})

        expect(screen.getByText("Default (always)")).toBeInTheDocument()
        expect(screen.getByText("Any (default)")).toBeInTheDocument()
    })

    it("sends the extended chain only when a fallback is added and Save is clicked", () => {
        const {mutate} = renderSettings({pullPolicy: [ALWAYS]})

        fireEvent.click(screen.getByText("Add fallback"))
        fireEvent.click(screen.getByText("Save"))

        expect(mutate).toHaveBeenCalledWith(
            {name: "runner-1", pullPolicy: {values: [ALWAYS, IF_NOT_PRESENT]}},
            expect.anything(),
        )
    })

    it("sends an empty values list when the whole chain is cleared", () => {
        const {mutate} = renderSettings({pullPolicy: [NEVER]})

        fireEvent.click(screen.getByText("✕"))
        fireEvent.click(screen.getByText("Save"))

        expect(mutate).toHaveBeenCalledWith(
            {name: "runner-1", pullPolicy: {values: []}},
            expect.anything(),
        )
    })

    it("sends allowedPullPolicies when an allowed policy is checked", () => {
        const {mutate} = renderSettings({})

        fireEvent.click(screen.getAllByRole("checkbox")[0])
        fireEvent.click(screen.getByText("Save"))

        expect(mutate).toHaveBeenCalledWith(
            {name: "runner-1", allowedPullPolicies: {values: [ALWAYS]}},
            expect.anything(),
        )
    })

    it("keeps Save disabled when the check interval is not a number", () => {
        renderSettings({})
        const checkInterval = screen.getAllByPlaceholderText("0 (default)")[0]

        fireEvent.change(checkInterval, {target: {value: "30"}})
        expect(screen.getByText("Save")).toBeEnabled()

        fireEvent.change(checkInterval, {target: {value: "abc"}})
        expect(screen.getByText("Save")).toBeDisabled()
    })
})
