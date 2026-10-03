import {afterEach, describe, expect, it, vi} from "vitest"

import {defaultBackendUrl} from "@/app/settings/defaults.ts"

describe("defaultBackendUrl", () => {
    afterEach(() => {
        vi.unstubAllEnvs()
    })

    it("falls back to the origin the UI is served from when no backend url is configured", () => {
        vi.stubEnv("VITE_VELEZ_BACKEND_URL", "")

        expect(defaultBackendUrl()).toBe(window.location.origin)
    })

    it("prefers the configured backend url", () => {
        vi.stubEnv("VITE_VELEZ_BACKEND_URL", "http://velez.internal:53891")

        expect(defaultBackendUrl()).toBe("http://velez.internal:53891")
    })
})
