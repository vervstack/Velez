import {describe, expect, it} from "vitest"
import {render, screen} from "@testing-library/react"

import Input from "@/components/base/Input.tsx"

describe("Input", () => {
    it("shows the error message and marks the field invalid when error is set", () => {
        render(<Input inputValue="Foo" onChange={() => undefined} error="Name is invalid"/>)

        expect(screen.getByText("Name is invalid")).toBeInTheDocument()
        expect(screen.getByRole("textbox")).toHaveAttribute("aria-invalid", "true")
    })

    it("renders no error and no aria-invalid when error is omitted", () => {
        render(<Input inputValue="foo" onChange={() => undefined}/>)

        expect(screen.queryByRole("alert")).not.toBeInTheDocument()
        expect(screen.getByRole("textbox")).not.toHaveAttribute("aria-invalid")
    })
})
