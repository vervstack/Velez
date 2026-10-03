import {describe, expect, it} from "vitest"
import {render, screen} from "@testing-library/react"

import RestartNotice from "@/dialogs/CreateServiceDialog/components/RestartNotice/RestartNotice.tsx"

describe("RestartNotice", () => {
    it("warns about a restart on a single-node setup", () => {
        render(<RestartNotice isClusterMode={false}/>)

        expect(screen.getByRole("note")).toHaveTextContent("The container will be recreated")
    })

    it("says the container stays running in cluster mode", () => {
        render(<RestartNotice isClusterMode/>)

        expect(screen.getByRole("note")).toHaveTextContent("The container stays running")
    })
})
