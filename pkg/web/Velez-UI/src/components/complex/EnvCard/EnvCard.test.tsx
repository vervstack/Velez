import {render, screen} from "@testing-library/react"
import {describe, expect, it} from "vitest"

import EnvCard from "@/components/complex/EnvCard/EnvCard.tsx"
import type {ServiceEnvironment} from "@/model/service_page/ServicePageModel"

function renderCard(overrides: Partial<ServiceEnvironment> = {}) {
    const env: ServiceEnvironment = {
        id: "prod",
        label: "prod",
        status: "running",
        version: "sha256:abcdef0123456789",
        deployedAgo: "1h ago",
        containerId: "0123456789abcdef0123456789abcdef",
        containerName: "s3_artel",
        health: "healthy",
        ...overrides,
    }

    return render(<EnvCard env={env}/>)
}

describe("EnvCard", () => {
    it("shows the short container id, name and image when they are known", () => {
        renderCard()

        expect(screen.getByText("0123456789ab")).toBeTruthy()
        expect(screen.getByText("s3_artel")).toBeTruthy()
        expect(screen.getByText("sha256:abcdef0123456789")).toBeTruthy()
    })

    it("hides the id and name rows when the container is unknown", () => {
        renderCard({containerId: "", containerName: ""})

        expect(screen.queryByText("id")).toBeNull()
        expect(screen.queryByText("name")).toBeNull()
        expect(screen.getByText("image")).toBeTruthy()
    })
})
