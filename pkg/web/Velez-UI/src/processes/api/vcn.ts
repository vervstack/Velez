import {VcnApi} from "@/app/api/velez/verv_closed_network.pb"
import type {ConnectServiceRequest, CreateVcnNetworkRequest} from "@/app/api/velez/verv_closed_network.pb"
import {ApiService} from "@/processes/ApiService.ts"

// ConnectService can wait up to 60 seconds for the sidecar task to finish.
const CONNECT_SERVICE_TIMEOUT_MS = 65_000
const CREATE_NETWORK_TIMEOUT_MS = 120_000
// DropNetwork can wait up to a minute for a remote node to remove Headscale.
const DROP_NETWORK_TIMEOUT_MS = 65_000

class VcnService extends ApiService {
    connectService(req: ConnectServiceRequest) {
        return this.mutate((initReq) => VcnApi.ConnectService(req, initReq), CONNECT_SERVICE_TIMEOUT_MS)
    }

    listNetworks() {
        return this.execute((initReq) => VcnApi.ListNetworks({}, initReq))
    }

    getNetwork(id: string) {
        return this.execute((initReq) => VcnApi.GetNetwork({id: encodeURIComponent(id)}, initReq))
    }

    createNetwork(req: CreateVcnNetworkRequest) {
        return this.mutate((initReq) => VcnApi.CreateNetwork(req, initReq), CREATE_NETWORK_TIMEOUT_MS)
    }

    dropNetwork(id: string) {
        return this.mutate((initReq) => VcnApi.DropNetwork({id}, initReq), DROP_NETWORK_TIMEOUT_MS)
    }

    issueApiKey(networkId: string) {
        return this.mutate((initReq) => VcnApi.IssueNetworkApiKey({networkId}, initReq))
    }

    expireApiKey(networkId: string, prefix: string) {
        return this.mutate((initReq) => VcnApi.ExpireNetworkApiKey({networkId, prefix}, initReq))
    }
}

export const vcnService = new VcnService()
