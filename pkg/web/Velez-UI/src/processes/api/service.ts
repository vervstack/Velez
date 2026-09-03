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
    GetVervonomiconRequest,
    DescriptorFile,
    ResourceReconciliation,
} from "@/app/api/velez"

import {ApiService} from "@/processes/ApiService.ts"
import type {ServiceAbout, ServiceMetrics, ServiceResource, ServiceGraphData, ServiceGraphNode, ServiceEnvironment, VervonomiconDocs, ResourceReconciliationStatus} from "@/model/service_page/ServicePageModel"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {mapResourceConnectionStatus} from "@/processes/vervonomicon.ts"

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

export function getResourceMeta(resourceType: string): { icon: string; color: string } {
    return RESOURCE_META[resourceType.toLowerCase()]
        ?? {icon: (resourceType.slice(0, 2) || '?').toUpperCase(), color: 'var(--fg-dim)'}
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
    const meta = getResourceMeta(type)
    return {
        name:   r.name ?? '',
        type,
        status: mapResourceStatus(r.status),
        icon:   meta.icon,
        color:  meta.color,
        reconciliation: 'unknown',
    }
}

function toResourceReconciliationStatus(r: ResourceReconciliation): ResourceReconciliationStatus {
    return {
        name: r.name ?? '',
        resourceType: r.resourceType ?? '',
        status: mapResourceConnectionStatus(r.status),
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

    async fetchVervonomicon(serviceName: string, environment?: string): Promise<VervonomiconDocs> {
        return this.execute(async (req) => {
            const payload: GetVervonomiconRequest = {serviceName, environment}
            const res = await ServiceApi.GetVervonomicon(payload, req)

            function decodeBase64(encoded: string | Uint8Array | undefined): string {
                if (!encoded) return ""
                let str: string
                if (typeof encoded === "string") {
                    str = encoded
                } else {
                    str = new TextDecoder().decode(encoded)
                }
                if (typeof str !== "string" || !str.match(/^[A-Za-z0-9+/=]*$/)) {
                    return str
                }
                try {
                    return decodeURIComponent(
                        atob(str)
                            .split("")
                            .map(c => `%${(`00${c.charCodeAt(0).toString(16)}`).slice(-2)}`)
                            .join("")
                    )
                } catch {
                    return str
                }
            }

            return {
                files: (res.raw ?? []).map((f: DescriptorFile) => ({
                    path: f.path ?? "",
                    content: decodeBase64(f.content),
                })),
                resolvedYaml: res.resolvedYaml ?? "",
                source: res.source ?? "VERVONOMICON_SOURCE_UNSPECIFIED",
                environment: res.environment ?? "",
                resourceStatuses: (res.resourceStatuses ?? []).map(toResourceReconciliationStatus),
            }
        })
    }
}

export const serviceService = new ServiceService()
