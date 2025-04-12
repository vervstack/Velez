import {getBackendUrl} from "@/app/store/Settings";

import {toServices} from "@/processes/ControlPlaneApi/mappings/Services";

import {ControlPlane, ListServicesRequest} from "@vervstack/velez";

const initReq = {pathPrefix: getBackendUrl()};

export function setBackendUrl(url: string) {
    initReq.pathPrefix = url
}

export async function ListServices() {
    const req: ListServicesRequest = {} as ListServicesRequest

    const services = await ControlPlane.ListServices(req, initReq);
    return toServices(services);
}
