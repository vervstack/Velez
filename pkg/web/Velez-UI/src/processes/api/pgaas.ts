import {
    PostgresAPI,
    ListPgInstancesRequest,
    ListPgInstancesResponse,
    CreatePgInstanceRequest,
    CreatePgInstanceResponse,
    DropPgInstanceRequest,
    EnablePgReplicationRequest,
    EnablePgReplicationResponse,
    ReplicatePgDatabaseRequest,
    ReplicatePgDatabaseResponse,
    GetPgInstanceCredentialsRequest,
    GetPgInstanceCredentialsResponse,
} from "@/app/api/velez"
import {ApiService} from "@/processes/ApiService.ts"

class PostgresService extends ApiService {
    async listPgInstances(req: ListPgInstancesRequest): Promise<ListPgInstancesResponse> {
        return this.execute((initReq) => PostgresAPI.ListPgInstances(req, initReq))
    }

    async createPgInstance(req: CreatePgInstanceRequest): Promise<CreatePgInstanceResponse> {
        return this.mutate((initReq) => PostgresAPI.CreatePgInstance(req, initReq))
    }

    async dropPgInstance(name: string): Promise<void> {
        return this.mutate((initReq) => {
            const payload: DropPgInstanceRequest = {name}
            return PostgresAPI.DropPgInstance(payload, initReq).then()
        })
    }

    async replicatePgDatabase(req: ReplicatePgDatabaseRequest): Promise<ReplicatePgDatabaseResponse> {
        return this.mutate((initReq) => PostgresAPI.ReplicatePgDatabase(req, initReq))
    }

    async enablePgReplication(req: EnablePgReplicationRequest): Promise<EnablePgReplicationResponse> {
        return this.mutate((initReq) => PostgresAPI.EnablePgReplication(req, initReq))
    }

    async getPgInstanceCredentials(name: string): Promise<GetPgInstanceCredentialsResponse> {
        return this.execute((initReq) => {
            const payload: GetPgInstanceCredentialsRequest = {name}
            return PostgresAPI.GetPgInstanceCredentials(payload, initReq)
        })
    }
}

export const pgaasService = new PostgresService()
