import {
    ControlPlaneAPI,
    ListPluginsRequest,
    EnablePluginRequest,
    EnablePluginResponse,
    VervPluginType,
    EnableStatefullCluster,
    EnableHeadscaleServer,
    EnableRegistry,
    ListNodesResponse,
    ListEnvironmentsResponse,
    CreateEnvironmentRequest,
    CreateEnvironmentResponse,
    UpdateEnvironmentRequest,
    UpdateEnvironmentResponse,
    DeleteEnvironmentRequest,
    ListRegistriesResponse,
    CreateRegistryRequest,
    CreateRegistryResponse,
    UpdateRegistryRequest,
    UpdateRegistryResponse,
    DeleteRegistryRequest,
    RegistryType,
} from "@/app/api/velez";

import {toServices} from "@/processes/mappings/services.ts";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";
import {ApiService} from "@/processes/ApiService.ts";

class ControlPlaneService extends ApiService {
    async listNodes(): Promise<ListNodesResponse> {
        return this.execute((req) => ControlPlaneAPI.ListNodes({}, req))
    }

    async listPlugins(): Promise<VervPlugin[]> {
        return this.execute(async (req) => {
            const list = await ControlPlaneAPI.ListPlugins({} as ListPluginsRequest, req)
            return toServices(list.plugins || [])
        })
    }

    async enablePlugin(pluginType: VervPluginType): Promise<void> {
        return this.mutate((req) => {
            const payload: EnablePluginRequest = {plugin: pluginType}
            return ControlPlaneAPI.EnablePlugin(payload, req).then()
        })
    }

    async listEnvironments(): Promise<ListEnvironmentsResponse> {
        return this.execute((req) => ControlPlaneAPI.ListEnvironments({}, req))
    }

    async createEnvironment(name: string, suffix?: string): Promise<CreateEnvironmentResponse> {
        return this.mutate((req) => {
            const payload: CreateEnvironmentRequest = {name, suffix}
            return ControlPlaneAPI.CreateEnvironment(payload, req)
        })
    }

    async updateEnvironment(id: string, name: string, suffix?: string): Promise<UpdateEnvironmentResponse> {
        return this.mutate((req) => {
            const payload: UpdateEnvironmentRequest = {id, name, suffix}
            return ControlPlaneAPI.UpdateEnvironment(payload, req)
        })
    }

    async deleteEnvironment(id: string): Promise<void> {
        return this.mutate((req) => {
            const payload: DeleteEnvironmentRequest = {id}
            return ControlPlaneAPI.DeleteEnvironment(payload, req).then()
        })
    }

    async listRegistries(): Promise<ListRegistriesResponse> {
        return this.execute((req) => ControlPlaneAPI.ListRegistries({}, req))
    }

    async createRegistry(
        name: string, type: RegistryType, url?: string, username?: string, secret?: string, isDefault?: boolean
    ): Promise<CreateRegistryResponse> {
        return this.mutate((req) => {
            const payload: CreateRegistryRequest = {name, type, url, username, secret, isDefault}
            return ControlPlaneAPI.CreateRegistry(payload, req)
        })
    }

    async updateRegistry(
        id: string, name: string, type?: RegistryType, url?: string, username?: string, secret?: string,
        isDefault?: boolean
    ): Promise<UpdateRegistryResponse> {
        return this.mutate((req) => {
            const payload: UpdateRegistryRequest = {id, name, type, url, username, secret, isDefault}
            return ControlPlaneAPI.UpdateRegistry(payload, req)
        })
    }

    async deleteRegistry(id: string): Promise<void> {
        return this.mutate((req) => {
            const payload: DeleteRegistryRequest = {id}
            return ControlPlaneAPI.DeleteRegistry(payload, req).then()
        })
    }

    async enableStatefullPgCluster(cluster: EnableStatefullCluster): Promise<EnablePluginResponse> {
        return this.mutate((req) => {
            const payload: EnablePluginRequest = {
                plugin: VervPluginType.statefull_pg,
                statefullCluster: cluster,
            }
            return ControlPlaneAPI.EnablePlugin(payload, req)
        })
    }

    async enableHeadscaleServer(config: EnableHeadscaleServer): Promise<void> {
        return this.mutate((req) => {
            const payload: EnablePluginRequest = {
                plugin: VervPluginType.headscale,
                headscaleServer: config,
            }
            return ControlPlaneAPI.EnablePlugin(payload, req).then()
        })
    }

    async enableRegistry(config: EnableRegistry): Promise<EnablePluginResponse> {
        return this.mutate((req) => {
            const payload: EnablePluginRequest = {
                plugin: VervPluginType.registry,
                registry: config,
            }
            return ControlPlaneAPI.EnablePlugin(payload, req)
        })
    }
}

export const controlPlaneService = new ControlPlaneService()
