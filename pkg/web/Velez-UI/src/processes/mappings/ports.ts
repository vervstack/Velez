export function formatPortMapping(containerPort?: number, hostPort?: number): string {
    if (!containerPort) return "-"
    if (!hostPort) return String(containerPort)
    return containerPort + " → " + hostPort
}
