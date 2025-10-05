import {ControlPlane, ListServicesRequest} from "@vervstack/velez";

import {toServices} from "@/processes/mappings/services.ts";

import {InitReq} from "@/app/settings/state.ts";

export async function ListServices(initReq: InitReq) {
    const req: ListServicesRequest = {} as ListServicesRequest

    const services = await ControlPlane.ListServices(req, initReq);
    return toServices(services);
}
