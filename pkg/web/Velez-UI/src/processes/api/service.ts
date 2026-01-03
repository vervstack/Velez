import {
    ServiceApi,
    VervAppService,
    GetServiceRequest,
    CreateDeployRequest,
    CreateSmerdRequest,
} from "@vervstack/velez"

import {GetInitReq} from "@/processes/api/api.ts";

export async function GetServiceByName(name: string): Promise<VervAppService> {
    const r: GetServiceRequest = {
        name: name
    }

    const r_1 = await ServiceApi.GetService(r, GetInitReq());
    if (!r_1.vervService) {
        throw new Error("ServiceNotFound");
    }
    return r_1.vervService;
}

export async function GetServiceById(id: string): Promise<VervAppService> {
    const r: GetServiceRequest = {
        id: id.toString()
    }

    const r_1 = await ServiceApi.GetService(r, GetInitReq());
    if (!r_1.vervService) {
        throw new Error("ServiceNotFound");
    }
    return r_1.vervService;
}

export async function CreateNewDeployment(serviceId: string, newReq: CreateSmerdRequest) {
    const req: CreateDeployRequest = {
        serviceId: serviceId,
        new: newReq,
    }
    return ServiceApi.CreateDeploy(req, GetInitReq())
}
