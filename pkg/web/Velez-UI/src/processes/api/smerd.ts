import {
    VelezAPI,
    ListSmerdsRequest,
    ListSmerdsResponse,
    DropSmerdResponse,
    DropSmerdRequest,
    UpgradeSmerdRequest,
    SearchImagesResponse,
    SearchImagesRequest,
    Smerd as ProtoSmerd,
} from "@/app/api/velez";
import {ApiService} from "@/processes/ApiService.ts";
import {CreateSmerdReq, Smerd, Port, Volume, toProto} from "@/model/smerds/Smerds.ts";

class SmerdService extends ApiService {
    async listSmerds(req: ListSmerdsRequest): Promise<ListSmerdsResponse> {
        return this.execute((initReq) => VelezAPI.ListSmerds(req, initReq))
    }

    async fetchSmerds(): Promise<ListSmerdsResponse> {
        return this.execute((initReq) => {
            const req: ListSmerdsRequest = {limit: 50}
            return VelezAPI.ListSmerds(req, initReq)
        })
    }

    async fetchSmerdsByServiceId(serviceName: string): Promise<ListSmerdsResponse> {
        return this.execute((initReq) => {
            const req: ListSmerdsRequest = {name: serviceName, limit: 10}
            return VelezAPI.ListSmerds(req, initReq)
        })
    }

    async fetchSmerd(name: string): Promise<ProtoSmerd> {
        return this.execute(async (initReq) => {
            const req: ListSmerdsRequest = {name, limit: 1}
            const res = await VelezAPI.ListSmerds(req, initReq)
            if (!res.smerds || res.smerds.length === 0) throw new Error("Smerd not found")
            return res.smerds[0]
        })
    }

    async getSmerd(name: string): Promise<Smerd> {
        return this.execute(async (initReq) => {
            const req: ListSmerdsRequest = {name, limit: 1}
            const res = await VelezAPI.ListSmerds(req, initReq)
            if (!res.smerds || res.smerds.length === 0) throw new Error("Smerd not found")
            const s = res.smerds[0]
            return {
                name: s.name,
                imageName: s.imageName,
                ports: (s.ports || []).map((v) => ({
                    servicePort: v.servicePortNumber,
                    exposedPort: v.exposedTo,
                } as Port)),
                volumes: (s.volumes || []).map((v) => ({
                    containerPath: v.containerPath,
                    virtualVolume: v.volumeName,
                } as Volume)),
            } as Smerd
        })
    }

    async listImages(name: string): Promise<SearchImagesResponse> {
        return this.execute((initReq) => {
            const req: SearchImagesRequest = {name}
            return VelezAPI.SearchImages(req, initReq)
        })
    }

    async deploySmerd(smerd: CreateSmerdReq): Promise<ProtoSmerd> {
        return this.mutate((initReq) => VelezAPI.CreateSmerd(toProto(smerd), initReq))
    }

    async dropSmerd(names: string[]): Promise<DropSmerdResponse> {
        return this.mutate((initReq) => {
            const req: DropSmerdRequest = {name: names}
            return VelezAPI.DropSmerd(req, initReq)
        })
    }

    async upgradeSmerd(name: string, image: string): Promise<void> {
        return this.mutate((initReq) => {
            const req: UpgradeSmerdRequest = {name, image}
            return VelezAPI.UpgradeSmerd(req, initReq).then()
        })
    }
}

export const smerdService = new SmerdService()
