import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {RunnerProvider, RunnerScope} from "@/app/api/velez"
import type {RunnerForm} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"
import RunnerFields from "@/dialogs/CreateServiceDialog/components/RunnerFields/RunnerFields.tsx"

const FORM: RunnerForm = {
    provider: RunnerProvider.GITHUB,
    scope: RunnerScope.REPO,
    target: "acme/app",
    baseUrl: "",
    labels: "",
    dockerImage: "",
    concurrent: "",
    accessToken: "",
    registrationToken: "",
}

function renderFields(form: RunnerForm, isRegistrationTokenFound = false) {
    const onChange = vi.fn()
    render(<RunnerFields form={form} isRegistrationTokenFound={isRegistrationTokenFound} onChange={onChange}/>)
    return {onChange}
}

describe("RunnerFields", () => {
    it("reports an edited target on top of the rest of the form", () => {
        const {onChange} = renderFields(FORM)

        fireEvent.change(screen.getByDisplayValue("acme/app"), {target: {value: "acme/other"}})

        expect(onChange).toHaveBeenCalledWith({...FORM, target: "acme/other"})
    })

    it("shows the gitlab-only fields only for gitlab", () => {
        renderFields({...FORM, provider: RunnerProvider.GITLAB})
        expect(screen.getByText("GitLab Base URL")).toBeInTheDocument()
        expect(screen.getByText("Concurrent jobs (optional)")).toBeInTheDocument()
    })

    it("hides the gitlab-only fields for github", () => {
        renderFields(FORM)

        expect(screen.queryByText("GitLab Base URL")).not.toBeInTheDocument()
        expect(screen.queryByText("Concurrent jobs (optional)")).not.toBeInTheDocument()
    })

    it("notes a found registration token only when one was found", () => {
        renderFields(FORM, true)

        expect(screen.getByText(/registration token was found in the container/)).toBeInTheDocument()
    })

    it("shows no note when no registration token was found", () => {
        renderFields(FORM)

        expect(screen.queryByText(/was found in the container/)).not.toBeInTheDocument()
    })
})
