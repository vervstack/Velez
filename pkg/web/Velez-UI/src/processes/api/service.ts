import {
    ServiceApi,
    VervAppService,
    GetServiceRequest,
    CreateDeployRequest,
    CreateSmerdRequest,
    ListServicesRequest,
    ListServicesResponse,
    ListDeploymentsRequest,
    ListDeploymentsResponse,
    StopServiceRequest,
    RestartServiceRequest,
} from "@/app/api/velez"

import {ApiService} from "@/processes/ApiService.ts"
import type {ServiceMetrics, ServiceResource, ServiceGraphData, ServiceEnvironment, VervonomiconDocs} from "@/model/service_page/ServicePageModel"

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
            const payload: CreateDeployRequest = {
                serviceName,
                new: newReq,
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

    // TODO: implement — requires GetServiceMetrics RPC in api/grpc/service_api.proto
    async fetchServiceMetrics(_serviceName: string): Promise<ServiceMetrics> {
        return {
            replicas: '3 / 3',
            uptime: '14d 3h 22m',
            cpu: 18,
            mem: 312,
            memMax: 1024,
            description: 'Multilevel YAML config sync. Pushes runtime config to subscribed services without redeploy.',
            type: 'gRPC service',
            team: 'platform',
            repo: 'godverv/matreshka',
            image: 'godverv/matreshka:0.4.7',
            port: '50051',
        }
    }

    // TODO: implement — requires GetServiceResources RPC in api/grpc/service_api.proto
    async fetchServiceResources(_serviceName: string): Promise<ServiceResource[]> {
        return [
            { id: 'redis-cache',    kind: 'redis',    icon: 'R', desc: 'cache + session',   host: 'redis.internal:6379',    status: 'healthy',  use: 'r/w',                  hits: '12.4k/s', color: '#ed2f32' },
            { id: 'kafka-events',   kind: 'kafka',    icon: 'K', desc: 'event stream',      host: 'kafka-1.internal:9092',  status: 'healthy',  use: 'producer · 4 topics',  hits: '820/s',   color: '#f5a623' },
            { id: 'postgres-main',  kind: 'postgres', icon: 'P', desc: 'primary store',     host: 'pg-main.internal:5432',  status: 'healthy',  use: 'r/w',                  hits: '2.1k qps',color: '#0ab7ee' },
            { id: 's3-blobs',       kind: 's3',       icon: 'S', desc: 'blob storage',      host: 's3.internal/matreshka',  status: 'healthy',  use: 'r/w',                  hits: '180/s',   color: '#a78bfa' },
            { id: 'elastic-search', kind: 'elastic',  icon: 'E', desc: 'search index',      host: 'es.internal:9200',       status: 'degraded', use: 'read',                 hits: '34/s',    color: '#28c840' },
        ]
    }

    // TODO: implement — requires GetServiceGraph RPC in api/grpc/service_api.proto
    async fetchServiceGraph(_serviceName: string): Promise<ServiceGraphData> {
        return {
            incoming: [
                { id: 'api-gateway',  kind: 'service', proto: 'gRPC', rate: '1.8k rps' },
                { id: 'worker-queue', kind: 'service', proto: 'gRPC', rate: '420 rps'  },
                { id: 'velez-ui',     kind: 'service', proto: 'HTTP', rate: '12 rps'   },
                { id: 'makosh',       kind: 'service', proto: 'gRPC', rate: '6 rps'    },
            ],
            outgoing: [
                { id: 'postgres-main', kind: 'resource', proto: 'pg',    rate: '2.1k qps' },
                { id: 'redis-cache',   kind: 'resource', proto: 'resp',  rate: '12k/s'    },
                { id: 'kafka-events',  kind: 'resource', proto: 'kafka', rate: '820/s'    },
                { id: 's3-blobs',      kind: 'resource', proto: 'http',  rate: '180/s'    },
                { id: 'svarog',        kind: 'service',  proto: 'gRPC',  rate: '40 rps'   },
            ],
        }
    }

    // TODO: implement — requires GetServiceEnvironments RPC in api/grpc/service_api.proto
    async fetchServiceEnvironments(_serviceName: string): Promise<ServiceEnvironment[]> {
        return [
            { id: 'dev',     label: 'dev',     status: 'running',  version: '0.4.9-rc.2', deployedAgo: '2h',  health: 'healthy'  },
            { id: 'test',    label: 'test',    status: 'running',  version: '0.4.8',      deployedAgo: '1d',  health: 'healthy'  },
            { id: 'staging', label: 'staging', status: 'degraded', version: '0.4.7',      deployedAgo: '5d',  health: 'degraded' },
            { id: 'prod',    label: 'prod',    status: 'running',  version: '0.4.7',      deployedAgo: '14d', health: 'healthy'  },
        ]
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
