import {describe, expect, it} from "vitest"
import {render, screen} from "@testing-library/react"

import {ValueOrigin} from "@/app/api/velez"
import PreviewRow from "@/dialogs/CreateServiceDialog/components/PreviewRow/PreviewRow.tsx"

describe("PreviewRow", () => {
    it("shows the plain value and the added tag for an added value", () => {
        render(<PreviewRow value={{key: "PORT", value: "8080", origin: ValueOrigin.ADDED}}/>)

        expect(screen.getByText("8080")).toBeInTheDocument()
        expect(screen.getByText("added")).toBeInTheDocument()
    })

    it("masks a secret value, never renders it and shows the hidden tag", () => {
        render(<PreviewRow value={{key: "TOKEN", value: "s3cr3t", isSecret: true}}/>)

        expect(screen.getByLabelText("hidden value")).toHaveTextContent("••••••")
        expect(screen.queryByText("s3cr3t")).not.toBeInTheDocument()
        expect(screen.getByText("hidden")).toBeInTheDocument()
    })

    it("shows only the hidden tag for a secret that is also added", () => {
        render(<PreviewRow value={{key: "TOKEN", value: "s3cr3t", isSecret: true, origin: ValueOrigin.ADDED}}/>)

        expect(screen.getByText("hidden")).toBeInTheDocument()
        expect(screen.queryByText("added")).not.toBeInTheDocument()
    })
})
