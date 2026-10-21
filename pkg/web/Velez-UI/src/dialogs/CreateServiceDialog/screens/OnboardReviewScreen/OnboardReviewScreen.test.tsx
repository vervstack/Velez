import {afterEach, describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import {AdoptGroup, useAdoptGroup} from "@/dialogs/CreateServiceDialog/hooks/AdoptGroup.ts"
import {
    initialOnboardContext,
    OnboardContext,
    OnboardFormState,
} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import OnboardReviewScreen from "@/dialogs/CreateServiceDialog/screens/OnboardReviewScreen/OnboardReviewScreen.tsx"

interface AdoptScreenProps {
    onFormChange(form: OnboardFormState): void
}

const FORM: OnboardFormState = {request: {containerId: "app"}, isRecreating: true, hasPublishedPorts: false}

function GroupProbe() {
    const group = useAdoptGroup()
    return <span>{`service ${group?.serviceName}`}</span>
}

vi.mock("@/dialogs/CreateServiceDialog/screens/GenericAdoptScreen/GenericAdoptScreen.tsx", () => ({
    default: ({onFormChange}: AdoptScreenProps) => (
        <button onClick={() => onFormChange(FORM)}>generic form <GroupProbe/></button>
    ),
}))
vi.mock("@/dialogs/CreateServiceDialog/screens/PostgresAdoptScreen/PostgresAdoptScreen.tsx", () => ({
    default: () => <span>postgres form</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/screens/RegistryAdoptScreen/RegistryAdoptScreen.tsx", () => ({
    default: () => <span>registry form</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/screens/RunnerAdoptScreen/RunnerAdoptScreen.tsx", () => ({
    default: ({initialProvider}: {initialProvider: string}) => <span>{`runner form ${initialProvider}`}</span>,
}))
vi.mock("@/dialogs/CreateServiceDialog/components/NetworkGroupList/NetworkGroupList.tsx", () => ({
    default: ({group, pattern}: {group: AdoptGroup, pattern: string}) => (
        <span>{`result of ${group.serviceName} as ${pattern}`}</span>
    ),
}))

const app = {id: "app", name: "app"}

function renderScreen(overrides: Partial<OnboardContext> = {}) {
    const updateContext = vi.fn()
    const context = {...initialOnboardContext(app, []), ...overrides}
    render(<OnboardReviewScreen {...context} updateContext={updateContext}/>)
    return {updateContext}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("OnboardReviewScreen", () => {
    it("renders the generic form for the chosen type with the result section below it", () => {
        renderScreen({screen: "generic"})

        expect(screen.getByText(/generic form/)).toBeInTheDocument()
        expect(screen.getByText("Result")).toBeInTheDocument()
        expect(screen.getByText("result of app as generic")).toBeInTheDocument()
    })

    it("hands the group of the context to the form", () => {
        renderScreen({screen: "generic", serviceNameDraft: "renamed"})

        expect(screen.getByText("service renamed")).toBeInTheDocument()
    })

    it("stores the form in the context when the form changes", () => {
        const {updateContext} = renderScreen({screen: "generic"})

        screen.getByRole("button").click()

        expect(updateContext).toHaveBeenCalledWith({form: FORM})
    })

    it("renders the runner form with the provider of the chosen type", () => {
        renderScreen({screen: "gitlabRunner"})

        expect(screen.getByText("runner form GITLAB")).toBeInTheDocument()
        expect(screen.getByText("result of app as runner")).toBeInTheDocument()
    })

    it("renders nothing while no type is chosen", () => {
        renderScreen({screen: undefined})

        expect(screen.queryByText("Result")).not.toBeInTheDocument()
    })
})
