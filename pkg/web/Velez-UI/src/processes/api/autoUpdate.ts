import {
    AutoUpdateAPI,
    ListAutoUpdateServicesResponse,
    SetServiceAutoUpdateRequest,
    SetServiceAutoUpdateResponse,
} from "@/app/api/velez/auto_update_api.pb"
import {ApiService} from "@/processes/ApiService.ts"

class AutoUpdateService extends ApiService {
    async listServices(): Promise<ListAutoUpdateServicesResponse> {
        return this.execute((initReq) => AutoUpdateAPI.ListAutoUpdateServices({}, initReq))
    }

    async setService(req: SetServiceAutoUpdateRequest): Promise<SetServiceAutoUpdateResponse> {
        return this.mutate((initReq) => AutoUpdateAPI.SetServiceAutoUpdate(req, initReq))
    }
}

export const autoUpdateService = new AutoUpdateService()
