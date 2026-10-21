import {describe, expect, it} from "vitest"
import {render, screen} from "@testing-library/react"

import TypeHint from "@/dialogs/CreateServiceDialog/components/TypeHint/TypeHint.tsx"

describe("TypeHint", () => {
    it("shows the text when visible", () => {
        render(<TypeHint isVisible text="Choose the type"/>)

        expect(screen.getByRole("status")).toHaveTextContent("Choose the type")
    })

    it("is hidden from the accessibility tree when not visible", () => {
        render(<TypeHint isVisible={false} text="Choose the type"/>)

        expect(screen.queryByRole("status")).not.toBeInTheDocument()
    })
})
