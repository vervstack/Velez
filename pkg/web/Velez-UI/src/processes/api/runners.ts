import {
    RunnersAPI,
    ListRunnersRequest,
    ListRunnersResponse,
    CreateRunnerRequest,
    CreateRunnerResponse,
    DropRunnerRequest,
} from "@/app/api/velez"
import {ApiService} from "@/processes/ApiService.ts"

class RunnersService extends ApiService {
    async listRunners(req: ListRunnersRequest): Promise<ListRunnersResponse> {
        return this.execute((initReq) => RunnersAPI.ListRunners(req, initReq))
    }

    async createRunner(req: CreateRunnerRequest): Promise<CreateRunnerResponse> {
        return this.mutate((initReq) => RunnersAPI.CreateRunner(req, initReq))
    }

    async dropRunner(name: string): Promise<void> {
        return this.mutate((initReq) => {
            const payload: DropRunnerRequest = {name}
            return RunnersAPI.DropRunner(payload, initReq).then()
        })
    }
}

export const runnersService = new RunnersService()
