import {LogStream, WatchContainerLogsResponse} from "@/app/api/velez"
import {WireTimestamp, wireTimestampToMillis} from "@/processes/mappings/wireTimestamp.ts"

export const MAX_LOG_LINES = 5000

export interface LogLine {
    id: number
    timestamp: string
    text: string
    isStderr: boolean
}

const ANSI_ESCAPE = new RegExp(String.fromCharCode(27) + "\\[[0-9;?]*[ -/]*[@-~]", "g")

export function stripAnsi(text: string): string {
    return text.replace(ANSI_ESCAPE, "")
}

export function formatLogTimestamp(ts?: WireTimestamp): string {
    const millis = wireTimestampToMillis(ts)
    if (!millis) return ""
    return new Date(millis).toISOString().replace("T", " ").replace("Z", "")
}

export function toLogLine(response: WatchContainerLogsResponse, id: number): LogLine {
    const text = stripAnsi(response.text || "").replace(/\r?\n$/, "")
    return {
        id,
        timestamp: formatLogTimestamp(response.timestamp),
        text,
        isStderr: response.stream === LogStream.STDERR,
    }
}

export function appendLogLines(existing: LogLine[], incoming: LogLine[], max: number = MAX_LOG_LINES): LogLine[] {
    const merged = existing.concat(incoming)
    return merged.length > max ? merged.slice(merged.length - max) : merged
}

export function filterLogLines(lines: LogLine[], query: string): LogLine[] {
    const needle = query.trim().toLowerCase()
    if (needle === "") return lines
    return lines.filter(function matchesQuery(line) {
        return line.text.toLowerCase().includes(needle)
    })
}

export interface QuerySegment {
    text: string
    isMatch: boolean
}

export function splitByQuery(text: string, query: string): QuerySegment[] {
    const needle = query.trim().toLowerCase()
    if (needle === "") return [{text, isMatch: false}]

    const haystack = text.toLowerCase()
    const segments: QuerySegment[] = []
    let cursor = 0
    let found = haystack.indexOf(needle, cursor)
    while (found !== -1) {
        if (found > cursor) segments.push({text: text.slice(cursor, found), isMatch: false})
        segments.push({text: text.slice(found, found + needle.length), isMatch: true})
        cursor = found + needle.length
        found = haystack.indexOf(needle, cursor)
    }
    if (cursor < text.length) segments.push({text: text.slice(cursor), isMatch: false})
    return segments
}
