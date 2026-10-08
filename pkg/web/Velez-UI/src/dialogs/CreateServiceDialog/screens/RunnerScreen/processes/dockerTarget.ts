export const EXTERNAL_DOCKER_CHOICE = "__external__"

export interface DockerTarget {
    dindName: string
    dockerSocketAddress: string
}

export function resolveDockerTarget(choice: string, externalAddress: string): DockerTarget {
    if (choice === EXTERNAL_DOCKER_CHOICE) {
        return {dindName: "", dockerSocketAddress: externalAddress.trim()}
    }
    return {dindName: choice, dockerSocketAddress: ""}
}

export function isBuildkitSupported(choice: string): boolean {
    return choice !== "" && choice !== EXTERNAL_DOCKER_CHOICE
}
