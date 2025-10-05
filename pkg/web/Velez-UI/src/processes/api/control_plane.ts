import {ControlPlane, ListServicesRequest} from "@vervstack/velez";

import {toServices} from "@/processes/mappings/services.ts";

import {InitReq} from "@/app/settings/state.ts";
import {Service} from "@/model/services/Services.tsx";

export async function ListServices(initReq: InitReq): Promise<Service[]> {
    const req: ListServicesRequest = {} as ListServicesRequest

    const services = await ControlPlane.ListServices(req, initReq);
    return toServices(services);
}
