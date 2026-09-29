export function isPgLoginMissing(env?: Record<string, string>): boolean {
    return !env?.POSTGRES_USER || !env?.POSTGRES_PASSWORD
}
