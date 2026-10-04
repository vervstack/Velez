import {DindAPI} from "@/app/api/velez/dind_api.pb"
import type {
    CreateDindRequest,
    CreateDindResponse,
    DropDindRequest,
    ListDindsResponse,
} from "@/app/api/velez/dind_api.pb"
import {ApiService} from "@/processes/ApiService.ts"

class DindsService extends ApiService {
    async listDinds(): Promise<ListDindsResponse> {
        return this.execute((initReq) => DindAPI.ListDinds({}, initReq))
    }

    async createDind(req: CreateDindRequest): Promise<CreateDindResponse> {
        return this.mutate((initReq) => DindAPI.CreateDind(req, initReq))
    }

    async dropDind(name: string): Promise<void> {
        return this.mutate((initReq) => {
            const payload: DropDindRequest = {name}
            return DindAPI.DropDind(payload, initReq).then()
        })
    }
}

export const dindsService = new DindsService()
