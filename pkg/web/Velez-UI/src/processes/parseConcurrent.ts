const WHOLE_NUMBER = /^\d+$/

export function parseConcurrent(value: string): number | undefined {
    const trimmed = value.trim()
    if (!WHOLE_NUMBER.test(trimmed)) return undefined

    const parsed = Number(trimmed)
    if (!Number.isSafeInteger(parsed) || parsed < 1) return undefined

    return parsed
}
