export type WireTimestamp = string | {seconds?: string | number, nanos?: number}

// grpc-gateway serialises google.protobuf.Timestamp as an RFC3339 string, the generated type says object
export function wireTimestampToMillis(ts?: WireTimestamp): number {
    if (!ts) return 0
    if (typeof ts === "string") {
        const parsed = new Date(ts).getTime()
        return Number.isNaN(parsed) ? 0 : parsed
    }
    return Number(ts.seconds || 0) * 1000 + Math.floor((ts.nanos || 0) / 1e6)
}
