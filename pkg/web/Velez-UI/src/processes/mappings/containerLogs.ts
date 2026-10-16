import {LogStream, WatchContainerLogsResponse} from "@/app/api/velez"

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

export function formatLogTimestamp(ts?: {seconds?: string | number, nanos?: number}): string {
    if (!ts) return ""
    const millis = Number(ts.seconds || 0) * 1000 + Math.floor((ts.nanos || 0) / 1e6)
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
