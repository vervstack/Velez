import {
    ServiceApi,
    VervAppService,
    GetServiceRequest,
    GetServiceMetricsRequest,
    GetServiceGraphRequest,
    ServiceDependencyInfo,
    CreateDeployRequest,
    CreateSmerdRequest,
    ListServicesRequest,
    ListServicesResponse,
    ListDeploymentsRequest,
    ListDeploymentsResponse,
    StopServiceRequest,
    RestartServiceRequest,
    RemoveServiceRequest,
    GetServiceEnvironmentsRequest,
    ServiceEnvironmentInfo,
    GetServiceResourcesRequest,
    BoundResource,
} from "@/app/api/velez"

import {ApiService} from "@/processes/ApiService.ts"
import type {ServiceAbout, ServiceMetrics, ServiceResource, ServiceGraphData, ServiceGraphNode, ServiceEnvironment, VervonomiconDocs} from "@/model/service_page/ServicePageModel"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"

function formatDeployedAgo(ts?: { seconds?: string | number; nanos?: number }): string {
    if (!ts?.seconds) return ''
    const seconds = Number(ts.seconds)
    if (!seconds) return ''
    const diffMs = Date.now() - seconds * 1000
    const diffDays = Math.floor(diffMs / 86400000)
    if (diffDays === 0) {
        const diffHours = Math.floor(diffMs / 3600000)
        if (diffHours === 0) {
            const diffMinutes = Math.floor(diffMs / 60000)
            return diffMinutes <= 1 ? 'just now' : `${diffMinutes}m`
        }
        return `${diffHours}h`
    }
    if (diffDays === 1) return '1d'
    return `${diffDays}d`
}

function mapEnvStatus(status?: string): ServiceEnvironment['status'] {
    switch ((status ?? '').toLowerCase()) {
        case 'running':
            return 'running'
        case 'degraded':
            return 'degraded'
        case 'failed':
            return 'failed'
        default:
            return 'stopped'
    }
}

function mapEnvHealth(health?: string): ServiceEnvironment['health'] {
    switch ((health ?? '').toLowerCase()) {
        case 'healthy':
            return 'healthy'
        case 'degraded':
            return 'degraded'
        default:
            return 'unhealthy'
    }
}

function toServiceEnvironment(info: ServiceEnvironmentInfo): ServiceEnvironment {
    return {
        id: info.env ?? '',
        label: info.env ?? '',
        status: mapEnvStatus(info.status),
        version: info.deployedVersion ?? '',
        deployedAgo: formatDeployedAgo(info.deployedAt),
        health: mapEnvHealth(info.health),
    }
}

// RESOURCE_META maps a backend resource_type to display-only chrome (icon glyph
// + accent color). Missing types fall back to the first two letters of the type.
const RESOURCE_META: Record<string, {icon: string; color: string}> = {
    postgres: {icon: 'Pg', color: 'var(--info-color)'},
    redis:    {icon: 'Rd', color: 'var(--red)'},
    kafka:    {icon: 'Kf', color: 'var(--amber)'},
    s3:       {icon: 'S3', color: 'var(--violet)'},
    mysql:    {icon: 'My', color: 'var(--info-color)'},
    mongo:    {icon: 'Mg', color: 'var(--green)'},
}

function mapResourceStatus(status?: string): ServiceResource['status'] {
    switch ((status ?? '').toLowerCase()) {
        case 'running':
        case 'healthy':
        case 'ok':
            return 'healthy'
        case 'degraded':
            return 'degraded'
        case '':
            return 'unknown'
        default:
            return 'unhealthy'
    }
}

function toServiceResource(r: BoundResource): ServiceResource {
    const type = r.resourceType ?? ''
    const meta = RESOURCE_META[type.toLowerCase()]
        ?? {icon: (type.slice(0, 2) || '?').toUpperCase(), color: 'var(--fg-dim)'}
    return {
        name:   r.name ?? '',
        type,
        status: mapResourceStatus(r.status),
        icon:   meta.icon,
        color:  meta.color,
    }
}

function formatUptime(seconds?: string): string {
    const total = Number(seconds ?? 0)
    if (total <= 0) return '—'
    const d = Math.floor(total / 86400)
    const h = Math.floor((total % 86400) / 3600)
    const m = Math.floor((total % 3600) / 60)
    const parts: string[] = []
    if (d > 0) parts.push(`${d}d`)
    if (h > 0) parts.push(`${h}h`)
    if (m > 0 || parts.length === 0) parts.push(`${m}m`)
    return parts.join(' ')
}

class ServiceService extends ApiService {
    async getServiceByName(name: string): Promise<VervAppService> {
        return this.execute(async (req) => {
            const payload: GetServiceRequest = {name}
            const res = await ServiceApi.GetService(payload, req)
            console.log(res)
            if (!res.vervService) throw new Error("ServiceNotFound")
            return res.vervService
        })
    }

    async fetchService(name: string): Promise<VervAppService> {
        return this.execute(async (req) => {
            const payload: GetServiceRequest = {name}
            const res = await ServiceApi.GetService(payload, req)
            if (!res.vervService) throw new Error("ServiceNotFound")
            return res.vervService
        })
    }

    async listServices(r: ListServicesRequest): Promise<ListServicesResponse> {
        return this.execute((req) => ServiceApi.ListServices(r, req))
    }

    async fetchDeploymentsByServiceName(serviceName: string): Promise<ListDeploymentsResponse> {
        return this.execute((req) => {
            const payload: ListDeploymentsRequest = {
                serviceName,
                paging: {limit: '10', offset: '0'},
            }
            return ServiceApi.ListDeployments(payload, req)
        })
    }

    async createNewDeployment(serviceName: string, newReq: CreateSmerdRequest): Promise<void> {
        return this.mutate((req) => {
            const environment = useEnvironmentStore.getState().selectedEnvironment
            const newReqWithEnvironment: CreateSmerdRequest = {
                ...newReq,
                environment: newReq.environment || environment,
            }
            const payload: CreateDeployRequest = {
                serviceName,
                environment,
                new: newReqWithEnvironment,
            }
            //  TODO remove
            payload.new!.imageName = 'redsockruf/zpotify'
            return ServiceApi.CreateDeploy(payload, req).then()
        })
    }

    async stopService(name: string): Promise<void> {
        return this.mutate((req) => {
            const payload: StopServiceRequest = {name}
            return ServiceApi.StopService(payload, req).then()
        })
    }

    async restartService(name: string): Promise<void> {
        return this.mutate((req) => {
            const payload: RestartServiceRequest = {name}
            return ServiceApi.RestartService(payload, req).then()
        })
    }

    async removeService(name: string, dropRunningInstances: boolean): Promise<void> {
        return this.mutate((req) => {
            const payload: RemoveServiceRequest = {name, dropRunningInstances}
            return ServiceApi.RemoveService(payload, req).then()
        })
    }

    async fetchServiceAbout(name: string): Promise<ServiceAbout> {
        return this.execute(async (req) => {
            const payload: GetServiceRequest = {name}
            const res = await ServiceApi.GetService(payload, req)
            const a = res.about
            return {
                description:  a?.description  ?? '',
                originalName: a?.originalName ?? '',
                env:          a?.env          ?? '',
                type:         a?.serviceType  ?? '',
                team:         a?.team         ?? '',
                repo:         a?.repo         ?? '',
                port:         a?.port         ?? '',
            }
        })
    }

    async fetchServiceMetrics(serviceName: string): Promise<ServiceMetrics> {
        return this.execute(async (req) => {
            const payload: GetServiceMetricsRequest = {serviceName}
            const res = await ServiceApi.GetServiceMetrics(payload, req)
            return {
                replicas: `${res.replicasRunning ?? 0} / ${res.replicasDesired ?? 0}`,
                uptime:   formatUptime(res.uptimeSeconds),
                cpu:      res.cpuPercent ?? 0,
                mem:      Number(res.memMi ?? 0),
                memMax:   Number(res.memMaxMi ?? 0),
            }
        })
    }

    async fetchServiceResources(serviceName: string): Promise<ServiceResource[]> {
        return this.execute(async (req) => {
            const payload: GetServiceResourcesRequest = {serviceName}
            const res = await ServiceApi.GetServiceResources(payload, req)
            return (res.resources ?? []).map(toServiceResource)
        })
    }

    async fetchServiceGraph(serviceName: string): Promise<ServiceGraphData> {
        return this.execute(async (req) => {
            const payload: GetServiceGraphRequest = { serviceName }
            const res = await ServiceApi.GetServiceGraph(payload, req)

            function toNode(dep: ServiceDependencyInfo): ServiceGraphNode {
                return {
                    id:    dep.serviceName ?? '',
                    kind:  dep.nodeType === 'NODE_TYPE_RESOURCE' ? 'resource' : 'service',
                    proto: dep.proto ?? '',
                    rate:  dep.requestRate != null ? `${dep.requestRate} rps` : '',
                }
            }

            return {
                incoming: (res.callers      ?? []).map(toNode),
                outgoing: (res.dependencies ?? []).map(toNode),
            }
        })
    }

    async fetchServiceEnvironments(serviceName: string): Promise<ServiceEnvironment[]> {
        return this.execute(async (req) => {
            const payload: GetServiceEnvironmentsRequest = {serviceName}
            const res = await ServiceApi.GetServiceEnvironments(payload, req)
            return (res.environments ?? []).map(toServiceEnvironment)
        })
    }

    // TODO: implement — requires GetVervonomicon RPC in api/grpc/service_api.proto
    async fetchVervonomicon(_serviceName: string): Promise<VervonomiconDocs> {
        return {
            vervonomicon: `# vervonomicon.yaml — service manifest
# Single source of truth for matreshka-be across all environments.

apiVersion: vervstack/v1
kind: Service
metadata:
  name: matreshka-be
  team: platform
  owner: r.popov@verv.dev
  repo: godverv/matreshka
  description: |
    Multilevel YAML config sync. Pushes runtime config
    to subscribed services without redeploy.

spec:
  type: grpc
  language: go
  runtime: docker
  image: godverv/matreshka:0.4.7
  port: 50051

  replicas:
    min: 2
    max: 6
    target_cpu: 70

  health:
    grpc: /grpc.health.v1.Health/Check
    interval: 10s
    timeout: 2s

  resources:
    - redis-cache       # session + hot config
    - kafka-events      # config-changed topic
    - postgres-main     # source of truth
    - s3-blobs          # encrypted config archives
    - elastic-search    # config search index

  observability:
    logs:    loki
    metrics: prometheus
    traces:  tempo
    errors:  sentry`,

            deployment: `# deployment.hcl — environment overrides
# Env-specific knobs. Inherits everything from vervonomicon.yaml.

deployment "prod" {
  strategy   = "rolling"
  max_surge  = 1
  max_unavailable = 0

  replicas {
    min = 3
    max = 12
    target_cpu = 65
  }

  node_selector = {
    region = "eu-west"
    tier   = "stable"
  }

  resources {
    cpu_request = "500m"
    cpu_limit   = "2000m"
    mem_request = "512Mi"
    mem_limit   = "2Gi"
  }

  guards {
    require_approval = true
    require_canary   = true
    freeze_window    = "fri 18:00 — mon 09:00 UTC"
  }
}

deployment "staging" {
  strategy  = "recreate"
  replicas { min = 1; max = 2 }
  guards   { require_canary = false }
}

deployment "dev" {
  strategy = "recreate"
  replicas { min = 1; max = 1 }
}`,

            configuration: `# configuration.yaml — runtime config
# Pushed to running pods via Matreshka itself. Hot-reloadable.

server:
  grpc_port: 50051
  http_port: 8080
  shutdown_grace: 15s

push:
  enabled: true
  interval: 5s
  batch_size: 64
  max_subscribers: 4096

cache:
  ttl: 30s
  max_entries: 50000
  eviction: lru

logging:
  level: info
  format: json
  sample_rate: 0.1

flags:
  enable_runtime_push: true
  enable_grpc_reflection: false
  enable_config_diff_audit: true
  experimental_async_writer: false`,

            secrets: `# secrets.yaml — references only, never values
# Real values resolved by Svarog at pod start.

secrets:
  - key: DB_PASSWORD
    ref: svarog://platform/matreshka/db_password
    rotated: 2026-04-22
    next_rotation: 2026-07-22

  - key: REDIS_AUTH
    ref: svarog://platform/matreshka/redis_auth
    rotated: 2026-04-12
    next_rotation: 2026-07-12

  - key: KAFKA_SASL_PASSWORD
    ref: svarog://platform/kafka/sasl_pw
    rotated: 2026-03-30
    next_rotation: 2026-06-30

  - key: S3_SECRET_KEY
    ref: svarog://platform/s3/access_secret
    rotated: 2026-04-01
    next_rotation: 2026-07-01

  - key: SENTRY_DSN
    ref: svarog://platform/matreshka/sentry_dsn
    rotated: 2025-12-15
    next_rotation: 2026-06-15

# 5 secrets · last sync: 4 minutes ago · all references valid`,
        }
    }
}

export const serviceService = new ServiceService()
