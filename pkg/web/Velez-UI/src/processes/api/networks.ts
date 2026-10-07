import {NetworkAPI} from "@/app/api/velez/network_api.pb"
import type {
    ConnectContainerRequest,
    CreateNetworkRequest,
    CreateNetworkResponse,
    DeleteNetworkRequest,
    DisconnectContainerRequest,
    GetNetworkStatusResponse,
    ListNetworksRequest,
    ListNetworksResponse,
} from "@/app/api/velez/network_api.pb"
import {ApiService} from "@/processes/ApiService.ts"

class NetworksService extends ApiService {
    async getStatus(): Promise<GetNetworkStatusResponse> {
        return this.execute((initReq) => NetworkAPI.GetNetworkStatus({}, initReq))
    }

    async listNetworks(environment: string, isForeignIncluded: boolean): Promise<ListNetworksResponse> {
        const req: ListNetworksRequest = {environment, isForeignIncluded}
        return this.execute((initReq) => NetworkAPI.ListNetworks(req, initReq))
    }

    async createNetwork(req: CreateNetworkRequest): Promise<CreateNetworkResponse> {
        return this.mutate((initReq) => NetworkAPI.CreateNetwork(req, initReq))
    }

    async deleteNetwork(req: DeleteNetworkRequest): Promise<void> {
        return this.mutate((initReq) => NetworkAPI.DeleteNetwork(req, initReq).then())
    }

    async connectContainer(req: ConnectContainerRequest): Promise<void> {
        return this.mutate((initReq) => NetworkAPI.ConnectContainer(req, initReq).then())
    }

    async disconnectContainer(req: DisconnectContainerRequest): Promise<void> {
        return this.mutate((initReq) => NetworkAPI.DisconnectContainer(req, initReq).then())
    }
}

export const networksService = new NetworksService()
