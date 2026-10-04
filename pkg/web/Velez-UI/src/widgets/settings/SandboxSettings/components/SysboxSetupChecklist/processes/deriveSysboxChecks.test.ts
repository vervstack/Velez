import {describe, expect, it} from "vitest"

import type {GetSysboxStatusResponse} from "@/app/api/velez/settings_api.pb"

import {
    deriveSysboxChecks,
    summarizeSysboxChecks,
} from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/processes/deriveSysboxChecks.ts"

const healthyStatus: GetSysboxStatusResponse = {
    osType: "linux",
    kernelVersion: "5.15.0-91-generic",
    dockerVersion: "26.1.0",
    isRootless: false,
    isSnap: false,
    isRuntimeRegistered: true,
    containersTotal: 4,
    containersOnSysbox: 1,
}

function stateOf(id: string, checks: ReturnType<typeof deriveSysboxChecks>) {
    return checks.find(check => check.id === id)?.state
}

describe("deriveSysboxChecks", () => {
    it("returns every step as unknown when nothing has loaded", () => {
        const checks = deriveSysboxChecks(undefined, undefined, undefined)

        expect(checks).toHaveLength(9)
        expect(checks.every(check => check.state === "unknown")).toBe(true)
    })

    it("passes every checkable step for a healthy host with the setting on and a passed smoke test", () => {
        const checks = deriveSysboxChecks(healthyStatus, {isPassed: true}, true)

        expect(summarizeSysboxChecks(checks)).toEqual({passed: 8, total: 8})
        expect(stateOf("dind", checks)).toBe("unknown")
    })

    const kernelCases = [
        {name: "passes on exactly 5.12", kernelVersion: "5.12.0", state: "pass"},
        {name: "passes on a newer major", kernelVersion: "6.1.0-13-amd64", state: "pass"},
        {name: "fails below 5.12", kernelVersion: "5.4.0-150-generic", state: "fail"},
        {name: "is unknown when unparsable", kernelVersion: "custom", state: "unknown"},
    ]
    for (const tc of kernelCases) {
        it(`kernel step ${tc.name}`, () => {
            const checks = deriveSysboxChecks({...healthyStatus, kernelVersion: tc.kernelVersion}, undefined, false)

            expect(stateOf("kernel", checks)).toBe(tc.state)
        })
    }

    it("hints at shiftfs for an old kernel", () => {
        const checks = deriveSysboxChecks({...healthyStatus, kernelVersion: "5.4.0"}, undefined, false)

        expect(checks.find(check => check.id === "kernel")?.hint).toBe("older kernels need shiftfs")
    })

    const rootlessHint = "Docker is running rootless"
    const snapHint = "Docker is installed as a snap"
    const dockerCases = [
        {name: "fails on rootless Docker", status: {isRootless: true}, state: "fail", hint: rootlessHint},
        {name: "fails on snap Docker", status: {isSnap: true}, state: "fail", hint: snapHint},
        {name: "passes on regular Docker", status: {}, state: "pass", hint: undefined},
    ]
    for (const tc of dockerCases) {
        it(`docker step ${tc.name}`, () => {
            const checks = deriveSysboxChecks({...healthyStatus, ...tc.status}, undefined, false)
            const docker = checks.find(check => check.id === "docker")

            expect(docker?.state).toBe(tc.state)
            expect(docker?.hint).toBe(tc.hint)
        })
    }

    it("fails the linux step on another OS and the install and runtime steps without the runtime", () => {
        const status = {...healthyStatus, osType: "darwin", isRuntimeRegistered: false}
        const checks = deriveSysboxChecks(status, undefined, false)

        expect([stateOf("linux", checks), stateOf("install", checks), stateOf("runtime", checks)])
            .toEqual(["fail", "fail", "fail"])
    })

    it("keeps the services and smoke steps unknown until a smoke test ran", () => {
        const checks = deriveSysboxChecks(healthyStatus, undefined, false)

        expect([stateOf("services", checks), stateOf("smoke", checks)]).toEqual(["unknown", "unknown"])
    })

    it("fails the services and smoke steps with the failure text when the smoke test failed", () => {
        const checks = deriveSysboxChecks(healthyStatus, {isPassed: false, failure: "runtime exited 125"}, false)

        expect([stateOf("services", checks), stateOf("smoke", checks)]).toEqual(["fail", "fail"])
        expect(checks.find(check => check.id === "smoke")?.hint).toBe("runtime exited 125")
    })

    it("fails the enable step while the setting is off", () => {
        const checks = deriveSysboxChecks(healthyStatus, undefined, false)

        expect(stateOf("enable", checks)).toBe("fail")
    })

    it("reports how many containers run under Sysbox when the setting is on", () => {
        const checks = deriveSysboxChecks(healthyStatus, undefined, true)

        expect(checks.find(check => check.id === "enable")?.hint)
            .toBe("1 of 4 containers run under Sysbox — recreate the rest to move them")
    })
})
