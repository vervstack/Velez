import {ConnectServiceRequest, ConnectServiceResponse, VcnApi} from "@/app/api/velez"
import {ApiService} from "@/processes/ApiService.ts"

// ConnectService can wait up to 60 seconds for the sidecar task to finish.
const CONNECT_SERVICE_TIMEOUT_MS = 65_000

class VcnService extends ApiService {
    connectService(req: ConnectServiceRequest): Promise<ConnectServiceResponse> {
        return this.mutate((initReq) => VcnApi.ConnectService(req, initReq), CONNECT_SERVICE_TIMEOUT_MS)
    }
}

export const vcnService = new VcnService()
