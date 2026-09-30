import {describe, expect, it} from "vitest"

import {isRegistryLoginRequired} from "@/dialogs/CreateServiceDialog/processes/registryLogin.ts"

describe("isRegistryLoginRequired", () => {
    it("requires a login when htpasswd auth is configured", () => {
        expect(isRegistryLoginRequired({REGISTRY_AUTH: "htpasswd"})).toBe(true)
        expect(isRegistryLoginRequired({REGISTRY_AUTH_HTPASSWD_PATH: "/auth/htpasswd"})).toBe(true)
    })

    it("requires no login when the env has no auth", () => {
        expect(isRegistryLoginRequired({REGISTRY_HTTP_ADDR: ":5000"})).toBe(false)
        expect(isRegistryLoginRequired({REGISTRY_AUTH: "none", REGISTRY_AUTH_HTPASSWD_PATH: ""})).toBe(false)
        expect(isRegistryLoginRequired(undefined)).toBe(false)
    })
})
