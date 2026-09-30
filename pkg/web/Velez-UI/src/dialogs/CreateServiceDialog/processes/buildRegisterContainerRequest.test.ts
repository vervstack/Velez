import {describe, expect, it} from "vitest"

import {PortProtocol} from "@/app/api/velez"
import {
    buildRegisterContainerRequest,
    RegisterContainerForm,
} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"

const PORT = {servicePortNumber: 5432, exposedTo: 15432, protocol: PortProtocol.tcp}

function newForm(overrides: Partial<RegisterContainerForm> = {}): RegisterContainerForm {
    return {
        containerId: "abc",
        environment: "",
        serviceName: " app ",
        pattern: "generic",
        links: [{source: "/srv/a", destination: "/data", volumeName: " app_data "}],
        isClusterMode: false,
        isKeepingPorts: false,
        ports: [],
        ...overrides,
    }
}

describe("buildRegisterContainerRequest", () => {
    it("builds a generic request with trimmed names and links", () => {
        const req = buildRegisterContainerRequest(newForm())

        expect(req).toEqual({
            containerId: "abc",
            environment: undefined,
            serviceName: "app",
            bindMountLinks: [{source: "/srv/a", volumeName: "app_data"}],
            keepPortMapping: false,
            ports: [],
            generic: {},
        })
    })

    it("builds a postgres request without a login when the env already has it", () => {
        const req = buildRegisterContainerRequest(newForm({pattern: "postgres"}))

        expect(req?.pg).toEqual({})
        expect(req?.generic).toBeUndefined()
    })

    it("carries the asked login in the pg pattern", () => {
        const req = buildRegisterContainerRequest(newForm({
            pattern: "postgres",
            pgLogin: {superuser: " admin ", password: "secret"},
        }))

        expect(req?.pg).toEqual({superuser: "admin", password: "secret"})
    })

    it("refuses a postgres request whose asked login is incomplete", () => {
        const req = buildRegisterContainerRequest(newForm({
            pattern: "postgres",
            pgLogin: {superuser: "admin", password: ""},
        }))

        expect(req).toBeNull()
    })

    it("builds an empty registry pattern when the container has no auth", () => {
        const req = buildRegisterContainerRequest(newForm({pattern: "registry"}))

        expect(req?.registry).toEqual({})
        expect(req?.generic).toBeUndefined()
    })

    it("carries the asked login in the registry pattern", () => {
        const req = buildRegisterContainerRequest(newForm({
            pattern: "registry",
            registryLogin: {username: " ci ", password: "secret"},
        }))

        expect(req?.registry).toEqual({username: "ci", password: "secret"})
    })

    it("refuses a registry request whose asked login is incomplete", () => {
        expect(buildRegisterContainerRequest(newForm({
            pattern: "registry",
            registryLogin: {username: "ci", password: ""},
        }))).toBeNull()
        expect(buildRegisterContainerRequest(newForm({
            pattern: "registry",
            registryLogin: {username: " ", password: "secret"},
        }))).toBeNull()
    })

    it("sends the listed ports when the mapping is not kept", () => {
        const req = buildRegisterContainerRequest(newForm({ports: [PORT]}))

        expect(req?.keepPortMapping).toBe(false)
        expect(req?.ports).toEqual([PORT])
    })

    it("keeps the mapping with no listed ports", () => {
        const req = buildRegisterContainerRequest(newForm({isKeepingPorts: true}))

        expect(req?.keepPortMapping).toBe(true)
        expect(req?.ports).toEqual([])
    })

    it("refuses keeping the mapping together with listed ports", () => {
        expect(buildRegisterContainerRequest(newForm({isKeepingPorts: true, ports: [PORT]}))).toBeNull()
    })

    it("omits the port fields in cluster mode", () => {
        const req = buildRegisterContainerRequest(newForm({isClusterMode: true, ports: [PORT]}))

        expect(req).not.toBeNull()
        expect(req).not.toHaveProperty("keepPortMapping")
        expect(req).not.toHaveProperty("ports")
    })

    it("refuses an empty service name, container id or link volume name", () => {
        expect(buildRegisterContainerRequest(newForm({serviceName: "  "}))).toBeNull()
        expect(buildRegisterContainerRequest(newForm({containerId: ""}))).toBeNull()
        expect(buildRegisterContainerRequest(newForm({
            links: [{source: "/srv/a", destination: "/data", volumeName: " "}],
        }))).toBeNull()
    })
})
