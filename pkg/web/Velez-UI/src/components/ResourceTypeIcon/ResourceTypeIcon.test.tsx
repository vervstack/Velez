import {describe, it, expect} from "vitest"
import {render, screen} from "@testing-library/react"

import ResourceTypeIcon from "@/components/ResourceTypeIcon/ResourceTypeIcon.tsx"

describe("ResourceTypeIcon", () => {
    it("renders an svg icon when the type is known", () => {
        const {container} = render(<ResourceTypeIcon type="dind" fallback="Di"/>)

        expect(container.getElementsByTagName("svg").length).toBe(1)
        expect(screen.queryByText("Di")).toBeNull()
    })

    it("renders the fallback text when the type is unknown", () => {
        const {container} = render(<ResourceTypeIcon type="mongo" fallback="Xy"/>)

        expect(screen.getByText("Xy")).toBeTruthy()
        expect(container.getElementsByTagName("svg").length).toBe(0)
    })

    it("renders an svg icon when the type is upper case", () => {
        const {container} = render(<ResourceTypeIcon type="DIND" fallback="Di"/>)

        expect(container.getElementsByTagName("svg").length).toBe(1)
    })
})
