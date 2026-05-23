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
    id: string
    kind: string            // "redis" | "kafka" | "postgres" | "s3" | "elastic"
    icon: string            // single letter abbreviation
    desc: string
    host: string
    status: 'healthy' | 'degraded' | 'unhealthy'
    use: string             // e.g. "r/w"
    hits: string            // e.g. "12.4k/s"
    color: string           // css color string
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
