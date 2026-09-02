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
}

export interface ServiceMetrics {
    replicas: string
    uptime: string
    cpu: number
    mem: number
    memMax: number
}

export interface ServiceResource {
    name: string                                              // resource_name from the backend
    type: string                                             // resource_type, e.g. "postgres"
    status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown'
    icon: string                                             // derived from type — display only
    color: string                                            // derived from type — css color / token
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

export interface VervonomiconDocs {
    vervonomicon: string
    deployment: string
    configuration: string
    secrets: string
}
