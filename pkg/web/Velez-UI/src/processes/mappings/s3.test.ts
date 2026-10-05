import {describe, expect, it} from "vitest"

import {
    formatAccessFlags,
    formatBytes,
    mapS3InstanceStatus,
    s3ServiceName,
    s3WebUiLink,
    sortS3InstancesByName,
} from "@/processes/mappings/s3.ts"

describe("formatBytes", () => {
    it("returns 0 B when the value is missing or zero", () => {
        expect(formatBytes(undefined)).toBe("0 B")
        expect(formatBytes("0")).toBe("0 B")
    })

    it("scales int64 strings into binary units", () => {
        expect(formatBytes("512")).toBe("512 B")
        expect(formatBytes("1536")).toBe("1.5 KiB")
        expect(formatBytes(String(5 * 1024 * 1024 * 1024))).toBe("5 GiB")
    })
})

describe("mapS3InstanceStatus", () => {
    it("maps container states to dot statuses", () => {
        expect(mapS3InstanceStatus("running")).toBe("running")
        expect(mapS3InstanceStatus("restarting")).toBe("degraded")
        expect(mapS3InstanceStatus("exited")).toBe("stopped")
    })
})

describe("sortS3InstancesByName", () => {
    it("sorts by name without mutating the input", () => {
        const input = [{name: "b"}, {name: "a"}]

        expect(sortS3InstancesByName(input).map((i) => i.name)).toEqual(["a", "b"])
        expect(input[0].name).toBe("b")
    })
})

describe("formatAccessFlags", () => {
    it("joins the granted flags and falls back to none", () => {
        expect(formatAccessFlags({isRead: true, isWrite: true})).toBe("read+write")
        expect(formatAccessFlags({})).toBe("none")
    })
})

describe("s3WebUiLink", () => {
    it("returns undefined when the web UI sidecar is off", () => {
        expect(s3WebUiLink({name: "s3"})).toBeUndefined()
    })

    it("builds a link on the current host when the web UI port is set", () => {
        expect(s3WebUiLink({name: "s3", webUiPort: 3909})).toMatch(/:3909$/)
    })
})

describe("s3ServiceName", () => {
    it("prefixes the instance name the way the backend names the service", () => {
        expect(s3ServiceName("garage1")).toBe("s3_garage1")
    })
})
