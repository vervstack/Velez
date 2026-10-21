export interface RootSwitchInfo {
    oldRootName: string
    newRootName: string
    isClusterMode: boolean
}

export function rootSwitchMessage({oldRootName, newRootName, isClusterMode}: RootSwitchInfo): string {
    const lines = [
        `${newRootName} becomes the network root and ${oldRootName} rejoins its network as a sidecar.`,
        "The group is recreated, so its containers are briefly unavailable.",
        `The published ports, networks and aliases of ${oldRootName} move to ${newRootName}.`,
        `The image of ${oldRootName} must be on the sidecar whitelist, otherwise onboarding fails.`,
    ]

    if (isClusterMode) {
        lines.push("In cluster mode the containers are recreated too; before, they were only bound.")
    }

    return lines.join(" ")
}
