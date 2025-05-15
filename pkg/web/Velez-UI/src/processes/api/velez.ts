import {getBackendUrl} from "@/app/store/Settings";

import {ListSmerdsRequest, VelezAPI} from "@vervstack/velez";

const initReq = {pathPrefix: getBackendUrl()};

export function setBackendUrl(url: string) {
    initReq.pathPrefix = url
}

export async function ListSmerds(req: ListSmerdsRequest) {
    req.limit = req.limit || 10
    return VelezAPI.ListSmerds(req, initReq)
}
