import {ServiceApi, CreateServiceRequest} from "@vervstack/velez";

import {InitReq} from "@/app/settings/state.ts";

export async function CreateService(initReq: InitReq, req: CreateServiceRequest): Promise<void> {
    ServiceApi.CreateService(req, initReq);
    return
}
