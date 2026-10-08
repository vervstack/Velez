import {describe, expect, it} from "vitest"
import {render, screen} from "@testing-library/react"

import InstanceCount from "@/components/InstanceCount/InstanceCount.tsx"

describe("InstanceCount", () => {
    it("shows the count and label when loaded", () => {
        render(<InstanceCount count={3} label="instances" isLoading={false}/>)

        expect(screen.getByText("3 instances")).toBeInTheDocument()
    })

    it("shows only the label and no number when loading", () => {
        render(<InstanceCount count={3} label="instances" isLoading/>)

        expect(screen.getByText("instances")).toBeInTheDocument()
        expect(screen.queryByText(/3/)).not.toBeInTheDocument()
    })
})
