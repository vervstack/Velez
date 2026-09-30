export function isRegistryLoginRequired(env?: Record<string, string>): boolean {
    return env?.REGISTRY_AUTH === "htpasswd" || Boolean(env?.REGISTRY_AUTH_HTPASSWD_PATH)
}
