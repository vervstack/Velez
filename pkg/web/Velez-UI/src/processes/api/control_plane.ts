import {
    ControlPlaneAPI,
    ListServicesRequest,
    EnableServiceRequest,
    VervServiceType,
    EnableStatefullCluster,
} from "@vervstack/velez";

import {toServices} from "@/processes/mappings/services.ts";

import {InitReq} from "@/app/settings/state.ts";
import {Service} from "@/model/services/Services.tsx";


export async function ListServices(initReq: InitReq): Promise<Service[]> {
    const req: ListServicesRequest = {} as ListServicesRequest

    const list = await ControlPlaneAPI.ListServices(req, initReq);
    return toServices(list.services || []);
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
