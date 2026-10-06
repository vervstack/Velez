import type {SmerdStatus} from "@/app/api/velez"

export interface ServiceEnvironment {
    id: string
    label: string
    status: 'running' | 'degraded' | 'stopped' | 'failed'
    version: string
    deployedAgo: string
    health: 'healthy' | 'degraded' | 'unhealthy'
}

export interface ServiceAbout {
    description: string
    originalName: string
    env: string
    type: string
    team: string
    repo: string
    port: string
    sidecars: ServiceSidecarView[]
}

export interface ServiceSidecarView {
    containerId: string
    name: string
    imageName: string
    status?: SmerdStatus
}

export interface ServiceMetrics {
    replicas: string
    uptime: string
    cpu: number
    mem: number
    memMax: number
}

export type ResourceConnectionStatus = 'already_connected' | 'must_provision' | 'unknown'

export interface ResourceReconciliationStatus {
    name: string                                              // resource_name from the backend
    resourceType: string                                     // resource_type, e.g. "postgres"
    status: ResourceConnectionStatus
}

export interface ServiceResource {
    name: string                                              // resource_name from the backend
    type: string                                             // resource_type, e.g. "postgres"
    status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown'
    icon: string                                             // derived from type — display only
    color: string                                            // derived from type — css color / token
    reconciliation: ResourceConnectionStatus                 // 'unknown' when reconciliation wasn't reported
    webUiPort?: number
    webUiHost?: string
}

export interface ServiceGraphNode {
    id: string
    kind: 'service' | 'resource'
    proto: string
    rate: string
}

export interface ServiceGraphData {
    incoming: ServiceGraphNode[]
    outgoing: ServiceGraphNode[]
}

export interface VervonomiconFile {
    path: string
    content: string
}

export interface VervonomiconDocs {
    files: VervonomiconFile[]
    resolvedYaml: string
    source: string
    environment: string
    resourceStatuses: ResourceReconciliationStatus[]
}
