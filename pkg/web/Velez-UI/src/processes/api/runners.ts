import {
    RunnersAPI,
    ListRunnersRequest,
    ListRunnersResponse,
    CreateRunnerRequest,
    CreateRunnerResponse,
    DropRunnerRequest,
    GetRunnerCredentialsRequest,
    GetRunnerCredentialsResponse,
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

    async getRunnerCredentials(name: string): Promise<GetRunnerCredentialsResponse> {
        return this.execute((initReq) => {
            const payload: GetRunnerCredentialsRequest = {name}
            return RunnersAPI.GetRunnerCredentials(payload, initReq)
        })
    }
}

export const runnersService = new RunnersService()
