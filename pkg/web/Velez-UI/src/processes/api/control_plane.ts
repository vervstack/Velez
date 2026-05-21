import {
    ControlPlaneAPI,
    ListPluginsRequest,
    EnablePluginRequest,
    VervPluginType,
    EnableStatefullCluster,
    ListNodesResponse,
    ListEnvironmentsResponse,
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

    async enableStatefullPgCluster(cluster: EnableStatefullCluster): Promise<void> {
        return this.mutate((req) => {
            const payload: EnablePluginRequest = {
                plugin: VervPluginType.statefull_pg,
                statefullCluster: cluster,
            }
            return ControlPlaneAPI.EnablePlugin(payload, req).then()
        })
    }
}

export const controlPlaneService = new ControlPlaneService()
