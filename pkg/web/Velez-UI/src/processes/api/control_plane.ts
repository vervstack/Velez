import {ControlPlane, ListServicesRequest} from "@vervstack/velez";

import {toServices} from "@/processes/mappings/services.ts";

import {InitReq} from "@/app/settings/state.ts";
import {Service} from "@/model/services/Services.tsx";

interface CpServices {
    active: Service[];
    inactive: Service[];
}

export async function ListServices(initReq: InitReq): Promise<CpServices> {
    const req: ListServicesRequest = {} as ListServicesRequest

    const list = await ControlPlane.ListServices(req, initReq);
    return {
        active: toServices(list.services || []),
        inactive: toServices(list.inactiveServices || []),
    };
}
