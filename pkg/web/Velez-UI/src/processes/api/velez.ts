import {ListSmerdsRequest, VelezAPI} from "@vervstack/velez";
import {InitReq} from "@/app/settings/state.ts";

export async function ListSmerds(req: ListSmerdsRequest, initReq: InitReq) {
    req.limit = req.limit || 10
    return VelezAPI.ListSmerds(req, initReq)
}
