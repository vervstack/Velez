export const NO_PORTS_TEXT = "This container publishes no ports, so no port mapping is created."

export function restartMessage(containerName: string | undefined, hasPublishedPorts = true): string {
    const base = `${containerName || "The container"} will be recreated under a new container id. ` +
        "The old one is kept until you finish onboarding."
    return hasPublishedPorts ? base : `${base} ${NO_PORTS_TEXT}`
}
