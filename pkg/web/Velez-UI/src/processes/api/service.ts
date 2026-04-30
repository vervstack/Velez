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
} from "@/app/api/velez"

import {GetInitReq} from "@/processes/api/api.ts";


export async function GetServiceByName(name: string): Promise<VervAppService> {
    const r: GetServiceRequest = {
        name: name
    }
    const initReq = GetInitReq()
    console.log(initReq)
    const r_1 = await ServiceApi.GetService(r, initReq);
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
    //  TODO remove
    req.new.imageName = 'redsockruf/zpotify'
    return ServiceApi.CreateDeploy(req, GetInitReq())
}

export async function ListServices(r: ListServicesRequest): Promise<ListServicesResponse> {
    return ServiceApi.ListServices(r, GetInitReq());
}

export async function FetchService(key: string): Promise<VervAppService> {
    const serviceId = Number(key)
    const req: GetServiceRequest = serviceId.toString() === key
        ? {id: key}
        : {name: key}

    const res = await ServiceApi.GetService(req, GetInitReq())
    if (!res.vervService) {
        throw new Error("ServiceNotFound")
    }
    return res.vervService
}

export async function FetchDeployments(serviceId: string): Promise<ListDeploymentsResponse> {
    const req: ListDeploymentsRequest = {
        serviceId,
        paging: {limit: '10', offset: '0'},
    }
    return ServiceApi.ListDeployments(req, GetInitReq())
}
