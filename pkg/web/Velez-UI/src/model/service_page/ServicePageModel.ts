export interface ServiceEnvironment {
    id: string
    label: string
    status: 'running' | 'degraded' | 'stopped' | 'failed'
    version: string
    deployedAgo: string
    health: 'healthy' | 'degraded' | 'unhealthy'
}

export interface ServiceMetrics {
    replicas: string        // e.g. "3 / 3"
    uptime: string          // e.g. "14d 3h 22m"
    cpu: number             // percent 0-100
    mem: number             // MiB used
    memMax: number          // MiB limit
    description: string
    type: string            // e.g. "gRPC service"
    team: string
    repo: string
    image: string
    port: string
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
