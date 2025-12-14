import {VcnApi, ConnectUserRequest} from "@vervstack/velez";

import {InitReq} from "@/app/settings/state.ts";

export async function ConnectUserToVcn(initReq: InitReq, key: string, username: string): Promise<void> {
    const req: ConnectUserRequest = {
        key: key,
        username: username,
    } as ConnectUserRequest

    VcnApi.ConnectUser(req, initReq);
    return
}
