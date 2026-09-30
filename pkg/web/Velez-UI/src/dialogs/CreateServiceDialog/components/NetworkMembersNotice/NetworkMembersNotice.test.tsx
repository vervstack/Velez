import {describe, expect, it} from "vitest"
import {render, screen} from "@testing-library/react"

import NetworkMembersNotice
    from "@/dialogs/CreateServiceDialog/components/NetworkMembersNotice/NetworkMembersNotice.tsx"

describe("NetworkMembersNotice", () => {
    it("lists each sidecar with its image", () => {
        render(
            <NetworkMembersNotice
                members={[
                    {id: "s1", name: "app", imageName: "nginx:1"},
                    {id: "s2", name: "worker", imageName: "redis:7"},
                ]}
            />
        )

        expect(screen.getByText("Also migrating with this container:")).toBeInTheDocument()
        expect(screen.getByText("app (nginx:1)")).toBeInTheDocument()
        expect(screen.getByText("worker (redis:7)")).toBeInTheDocument()
    })
})
