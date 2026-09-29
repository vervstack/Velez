import {
    RunnersAPI,
    ListRunnersRequest,
    ListRunnersResponse,
    CreateRunnerRequest,
    CreateRunnerResponse,
    DropRunnerRequest,
    ReregisterRunnerRequest,
    ReregisterRunnerResponse,
    GetRunnerCredentialsRequest,
    GetRunnerCredentialsResponse,
    GetRunnerConfigRequest,
    GetRunnerConfigResponse,
    UpdateRunnerConfigRequest,
    UpdateRunnerConfigResponse,
    RedeployRunnerRequest,
    RedeployRunnerResponse,
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

    async reregisterRunner(name: string): Promise<ReregisterRunnerResponse> {
        return this.mutate((initReq) => {
            const payload: ReregisterRunnerRequest = {name}
            return RunnersAPI.ReregisterRunner(payload, initReq)
        })
    }

    async getRunnerCredentials(name: string): Promise<GetRunnerCredentialsResponse> {
        return this.execute((initReq) => {
            const payload: GetRunnerCredentialsRequest = {name}
            return RunnersAPI.GetRunnerCredentials(payload, initReq)
        })
    }

    async getRunnerConfig(name: string): Promise<GetRunnerConfigResponse> {
        return this.execute((initReq) => {
            const payload: GetRunnerConfigRequest = {name}
            return RunnersAPI.GetRunnerConfig(payload, initReq)
        })
    }

    async updateRunnerConfig(req: UpdateRunnerConfigRequest): Promise<UpdateRunnerConfigResponse> {
        return this.mutate((initReq) => RunnersAPI.UpdateRunnerConfig(req, initReq))
    }

    async redeployRunner(name: string): Promise<RedeployRunnerResponse> {
        return this.mutate((initReq) => {
            const payload: RedeployRunnerRequest = {name}
            return RunnersAPI.RedeployRunner(payload, initReq)
        })
    }
}

export const runnersService = new RunnersService()
