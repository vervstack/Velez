import {ServiceApi, VervAppService, GetServiceRequest} from "@vervstack/velez"
import {InitReq} from "@/app/settings/state.ts";

export async function GetServiceByName(initReq: InitReq, name: string): Promise<VervAppService> {
    const r: GetServiceRequest = {
        name: name
    }

    const r_1 = await ServiceApi.GetService(r, initReq);
    if (!r_1.vervService) {
        throw new Error("ServiceNotFound");
    }
    return r_1.vervService;
}


export async function GetServiceById(initReq: InitReq, id: string): Promise<VervAppService> {
    const r: GetServiceRequest = {
        id: id.toString()
    }

    const r_1 = await ServiceApi.GetService(r, initReq);
    if (!r_1.vervService) {
        throw new Error("ServiceNotFound");
    }
    return r_1.vervService;
}
