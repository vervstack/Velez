export const TYPE_DELAY_MS = 30
export const DELETE_DELAY_MS = 20

// Deletes down to the longest prefix shared with the target, then types toward it.
export function nextTypewriterFrame(shown: string, target: string): string {
    if (shown === target) {
        return shown
    }

    if (target.startsWith(shown)) {
        return target.slice(0, shown.length + 1)
    }

    return shown.slice(0, -1)
}

export function frameDelayMs(shown: string, next: string): number {
    return next.length < shown.length ? DELETE_DELAY_MS : TYPE_DELAY_MS
}
