import {
    ControlPlaneAPI,
    ListPluginsRequest,
    EnableServiceRequest,
    VervServiceType,
    EnableStatefullCluster, ListNodesResponse, ListNodesRequest,
} from "@/app/api/velez";

import {toServices} from "@/processes/mappings/services.ts";

import {InitReq} from "@/app/settings/state.ts";
import {Service} from "@/model/services/Services.tsx";
import {GetInitReq} from "@/processes/api/api.ts";

export async function ListNodes(): Promise<ListNodesResponse> {
    const req: ListNodesRequest = {}

    return ControlPlaneAPI.ListNodes(req, GetInitReq())
}

export async function ListPlugins(initReq: InitReq): Promise<Service[]> {
    const req: ListPluginsRequest = {} as ListPluginsRequest

    const list = await ControlPlaneAPI.ListPlugins(req, initReq);
    return toServices(list.plugins || []);
}


export async function EnableService(vervService: VervServiceType, initReq: InitReq): Promise<void> {
    const req: EnableServiceRequest = {
        service: vervService,
    }

    return ControlPlaneAPI.EnableService(req, initReq).then();
}

export async function EnableStatefullPgCluster(payload: EnableStatefullCluster, initReq: InitReq): Promise<void> {
    const req: EnableServiceRequest = {
        service: VervServiceType.statefull_pg,
        statefullCluster: payload
    }

    return ControlPlaneAPI.EnableService(req, initReq).then();
}
