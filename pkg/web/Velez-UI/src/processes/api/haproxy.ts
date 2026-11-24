import {HaproxyAPI} from "@/app/api/velez/haproxy_api.pb"
import type {
    CreateHaproxyInstanceRequest,
    DropHaproxyInstanceRequest,
    SelectHaproxyMasterRequest,
} from "@/app/api/velez/haproxy_api.pb"
import {ApiService} from "@/processes/ApiService.ts"

class HaproxyService extends ApiService {
    listHaproxyInstances() {
        return this.execute((initReq) => HaproxyAPI.ListHaproxyInstances({}, initReq))
    }

    createHaproxyInstance(req: CreateHaproxyInstanceRequest) {
        return this.mutate((initReq) => HaproxyAPI.CreateHaproxyInstance(req, initReq))
    }

    dropHaproxyInstance(name: string) {
        const req: DropHaproxyInstanceRequest = {name}
        return this.mutate((initReq) => HaproxyAPI.DropHaproxyInstance(req, initReq))
    }

    selectHaproxyMaster(req: SelectHaproxyMasterRequest) {
        return this.mutate((initReq) => HaproxyAPI.SelectHaproxyMaster(req, initReq))
    }
}

export const haproxyService = new HaproxyService()
