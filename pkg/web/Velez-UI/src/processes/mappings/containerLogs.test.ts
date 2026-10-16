import {describe, expect, it} from "vitest"

import {LogStream} from "@/app/api/velez"
import {
    appendLogLines,
    filterLogLines,
    formatLogTimestamp,
    LogLine,
    stripAnsi,
    toLogLine,
} from "@/processes/mappings/containerLogs.ts"

function newLine(id: number, text: string): LogLine {
    return {id, timestamp: "", text, isStderr: false}
}

describe("appendLogLines", () => {
    it("drops the oldest lines when the cap is exceeded", () => {
        const existing = [newLine(1, "a"), newLine(2, "b"), newLine(3, "c")]
        const result = appendLogLines(existing, [newLine(4, "d"), newLine(5, "e")], 4)
        expect(result.map(l => l.id)).toEqual([2, 3, 4, 5])
    })

    it("keeps every line when the cap is not exceeded", () => {
        const result = appendLogLines([newLine(1, "a")], [newLine(2, "b")], 4)
        expect(result.map(l => l.id)).toEqual([1, 2])
    })
})

describe("filterLogLines", () => {
    it("matches the query case-insensitively", () => {
        const lines = [newLine(1, "Connection ERROR"), newLine(2, "all good")]
        expect(filterLogLines(lines, "error").map(l => l.id)).toEqual([1])
    })

    it("returns every line when the query is blank", () => {
        const lines = [newLine(1, "a"), newLine(2, "b")]
        expect(filterLogLines(lines, "  ")).toEqual(lines)
    })
})

describe("stripAnsi", () => {
    it("removes colour escape sequences", () => {
        expect(stripAnsi("\u001b[31mfail\u001b[0m ok")).toBe("fail ok")
    })
})

describe("formatLogTimestamp", () => {
    it("renders seconds and nanos as a UTC date-time with milliseconds", () => {
        expect(formatLogTimestamp({seconds: "1700000000", nanos: 123000000})).toBe("2023-11-14 22:13:20.123")
    })

    it("returns an empty string when the timestamp is missing", () => {
        expect(formatLogTimestamp(undefined)).toBe("")
    })
})

describe("toLogLine", () => {
    it("marks stderr lines and strips ansi codes and the trailing newline", () => {
        const line = toLogLine({stream: LogStream.STDERR, text: "\u001b[31mboom\u001b[0m\n"}, 7)
        expect(line).toEqual({id: 7, timestamp: "", text: "boom", isStderr: true})
    })
})
