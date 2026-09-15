import {
    ContainerRegistryAPI,
    ListRegistryInstancesRequest,
    ListRegistryInstancesResponse,
    CreateRegistryInstanceRequest,
    CreateRegistryInstanceResponse,
    DropRegistryInstanceRequest,
    GetRegistryInstanceCredentialsRequest,
    GetRegistryInstanceCredentialsResponse,
} from "@/app/api/velez"
import {ApiService} from "@/processes/ApiService.ts"

class ContainerRegistryService extends ApiService {
    async listRegistryInstances(req: ListRegistryInstancesRequest): Promise<ListRegistryInstancesResponse> {
        return this.execute((initReq) => ContainerRegistryAPI.ListRegistryInstances(req, initReq))
    }

    async createRegistryInstance(req: CreateRegistryInstanceRequest): Promise<CreateRegistryInstanceResponse> {
        return this.mutate((initReq) => ContainerRegistryAPI.CreateRegistryInstance(req, initReq))
    }

    async dropRegistryInstance(name: string): Promise<void> {
        return this.mutate((initReq) => {
            const payload: DropRegistryInstanceRequest = {name}
            return ContainerRegistryAPI.DropRegistryInstance(payload, initReq).then()
        })
    }

    async getRegistryInstanceCredentials(name: string): Promise<GetRegistryInstanceCredentialsResponse> {
        return this.execute((initReq) => {
            const payload: GetRegistryInstanceCredentialsRequest = {name}
            return ContainerRegistryAPI.GetRegistryInstanceCredentials(payload, initReq)
        })
    }
}

export const registryaasService = new ContainerRegistryService()
