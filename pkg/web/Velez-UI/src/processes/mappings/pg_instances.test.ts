import {describe, it, expect} from "vitest"

import {
    mapPgInstanceStatus,
    formatPgInstanceCreatedAt,
    sortPgInstancesByName,
} from "@/processes/mappings/pg_instances.ts"

describe("pg_instances mappings", () => {
    describe("mapPgInstanceStatus", () => {
        it("maps a running status", () => {
            expect(mapPgInstanceStatus("running")).toBe("running")
        })

        it("maps a restarting status to degraded", () => {
            expect(mapPgInstanceStatus("restarting")).toBe("degraded")
        })

        it("maps an exited status to stopped", () => {
            expect(mapPgInstanceStatus("exited")).toBe("stopped")
        })

        it("defaults to stopped when status is missing", () => {
            expect(mapPgInstanceStatus(undefined)).toBe("stopped")
        })
    })

    describe("formatPgInstanceCreatedAt", () => {
        it("formats a timestamp with seconds", () => {
            const ts = {seconds: "1700000000"}
            expect(formatPgInstanceCreatedAt(ts)).toBe(new Date(1700000000 * 1000).toLocaleDateString())
        })

        it("returns a dash when seconds are missing", () => {
            expect(formatPgInstanceCreatedAt(undefined)).toBe("-")
        })

        it("returns a dash when seconds are zero", () => {
            expect(formatPgInstanceCreatedAt({seconds: "0"})).toBe("-")
        })
    })

    describe("sortPgInstancesByName", () => {
        it("sorts instances alphabetically by name", () => {
            const instances = [{name: "zebra"}, {name: "apple"}, {name: "mango"}]
            const sorted = sortPgInstancesByName(instances)
            expect(sorted.map((i) => i.name)).toEqual(["apple", "mango", "zebra"])
        })

        it("does not mutate the input array", () => {
            const instances = [{name: "b"}, {name: "a"}]
            sortPgInstancesByName(instances)
            expect(instances.map((i) => i.name)).toEqual(["b", "a"])
        })
    })
})
