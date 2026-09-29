const KEEPING_PORTS_TEXT = "It is stopped first to keep its exact host ports, so it stays down until the new one is up."
const DEFAULT_TEXT = "It is unavailable while the new one starts."

export function restartMessage(containerName: string | undefined, isKeepingPorts: boolean): string {
    const downtime = isKeepingPorts ? KEEPING_PORTS_TEXT : DEFAULT_TEXT
    return `${containerName || "The container"} will be recreated under a new container id. ${downtime}`
}
