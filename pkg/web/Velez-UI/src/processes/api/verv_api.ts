import {ServiceApi, CreateServiceRequest} from "@/app/api/velez";

import {InitReq} from "@/app/settings/state.ts";
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts";

export async function CreateService(initReq: InitReq, req: CreateServiceRequest): Promise<void> {
    req.environment = req.environment || useEnvironmentStore.getState().selectedEnvironment
    ServiceApi.CreateService(req, initReq);
    return
}
