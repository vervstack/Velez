import type {ResourceConnectionStatus, ResourceReconciliationStatus, ServiceResource} from "@/model/service_page/ServicePageModel"

export interface VervonomiconTab {
    id: string
    label: string
    filePath: string
}

export function buildVervonomiconTabs(filePaths: string[]): VervonomiconTab[] {
    return filePaths.map(path => ({
        id: path,
        label: path,
        filePath: path,
    }))
}

export function findDefaultTab(tabs: VervonomiconTab[]): string {
    const vervonomicon = tabs.find(t => t.filePath === "vervonomicon.yaml")
    return vervonomicon ? vervonomicon.id : (tabs[0]?.id ?? "")
}

export function isYamlFile(filePath: string): boolean {
    return /\.(yaml|yml)$/.test(filePath)
}

export function shouldHighlightYaml(filePath: string): boolean {
    return isYamlFile(filePath)
}

// mapResourceConnectionStatus maps the proto ResourceConnectionStatus enum string to an
// app-level union, decoupled from the generated proto names.
export function mapResourceConnectionStatus(protoStatus?: string): ResourceConnectionStatus {
    switch (protoStatus) {
        case "RESOURCE_CONNECTION_STATUS_ALREADY_CONNECTED":
            return "already_connected"
        case "RESOURCE_CONNECTION_STATUS_MUST_PROVISION":
            return "must_provision"
        default:
            return "unknown"
    }
}

// canCreateResource says whether the UI may offer to create/provision a resource.
// A resource the service is already connected to must never be re-created.
export function canCreateResource(status: ResourceConnectionStatus): boolean {
    return status !== "already_connected"
}

// mergeResourcesWithReconciliation overlays reconciliation status onto the bound resources
// returned by GetServiceResources, and appends a placeholder entry for every resource declared
// in resources.yaml that isn't bound yet (status "must_provision") so the page can show it as
// declared-but-not-yet-created.
export function mergeResourcesWithReconciliation(
    resources: ServiceResource[],
    reconciliation: ResourceReconciliationStatus[],
    getMeta: (resourceType: string) => { icon: string; color: string },
): ServiceResource[] {
    const reconciliationByName = new Map(reconciliation.map(r => [r.name, r]))

    const merged = resources.map(resource => {
        const match = reconciliationByName.get(resource.name)
        return match ? {...resource, reconciliation: match.status} : resource
    })

    const boundNames = new Set(resources.map(r => r.name))
    const unprovisioned = reconciliation
        .filter(r => r.status === "must_provision" && !boundNames.has(r.name))
        .map(r => {
            const meta = getMeta(r.resourceType)
            const placeholder: ServiceResource = {
                name: r.name,
                type: r.resourceType,
                status: "unknown",
                icon: meta.icon,
                color: meta.color,
                reconciliation: r.status,
            }
            return placeholder
        })

    return [...merged, ...unprovisioned]
}
