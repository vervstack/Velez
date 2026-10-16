import {describe, expect, it} from "vitest"

import {LogStream} from "@/app/api/velez"
import {
    appendLogLines,
    filterLogLines,
    formatLogTimestamp,
    LogLine,
    splitByQuery,
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

    it("renders the RFC3339 string grpc-gateway sends", () => {
        expect(formatLogTimestamp("2026-10-16T14:22:01.123456789Z")).toBe("2026-10-16 14:22:01.123")
    })

    it("returns an empty string when the timestamp is missing", () => {
        expect(formatLogTimestamp(undefined)).toBe("")
    })

    it("returns an empty string when the string is not a date", () => {
        expect(formatLogTimestamp("garbage")).toBe("")
    })
})

describe("splitByQuery", () => {
    it("returns one non-match segment for a blank query", () => {
        expect(splitByQuery("hello", "  ")).toEqual([{text: "hello", isMatch: false}])
    })

    it("marks every case-insensitive occurrence and keeps the original casing", () => {
        expect(splitByQuery("Err ok err", "ERR")).toEqual([
            {text: "Err", isMatch: true},
            {text: " ok ", isMatch: false},
            {text: "err", isMatch: true},
        ])
    })

    it("treats regex characters in the query literally", () => {
        expect(splitByQuery("a.c abc", ".")).toEqual([
            {text: "a", isMatch: false},
            {text: ".", isMatch: true},
            {text: "c abc", isMatch: false},
        ])
    })

    it("returns one non-match segment when nothing matches", () => {
        expect(splitByQuery("abc", "x")).toEqual([{text: "abc", isMatch: false}])
    })
})

describe("toLogLine", () => {
    it("marks stderr lines and strips ansi codes and the trailing newline", () => {
        const line = toLogLine({stream: LogStream.STDERR, text: "\u001b[31mboom\u001b[0m\n"}, 7)
        expect(line).toEqual({id: 7, timestamp: "", text: "boom", isStderr: true})
    })
})
