import {
    VelezAPI,
    TasksApi,
    GetHardwareResponse,
    ListSmerdsRequest,
    ListSmerdsResponse,
    SearchImagesRequest,
    SearchImagesResponse,
    Smerd as ProtoSmerd,
    TaskStatus,
    VersionResponse,
} from "@/app/api/velez";
import {InitReq} from "@/app/settings/state.ts";
import {CreateSmerdReq, Port, Smerd, toProto, Volume} from "@/model/smerds/Smerds.ts";
import {ApiService} from "@/processes/ApiService.ts";
import {GetInitReq} from "@/processes/api/api.ts";
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts";

class VelezService extends ApiService {
    async ping(): Promise<VersionResponse> {
        return this.mutate((req) => VelezAPI.Version({}, req))
    }
}

export const velezService = new VelezService()

export async function ListSmerds(req: ListSmerdsRequest, initReq: InitReq) {
    req.limit = req.limit || 10
    req.environment = req.environment || useEnvironmentStore.getState().selectedEnvironment
    return VelezAPI.ListSmerds(req, initReq)
}

export async function FetchSmerds(): Promise<ListSmerdsResponse> {
    const req: ListSmerdsRequest = {limit: 50, environment: useEnvironmentStore.getState().selectedEnvironment}
    return VelezAPI.ListSmerds(req, GetInitReq())
}

export async function FetchSmerdsByServiceId(serviceName: string): Promise<ListSmerdsResponse> {
    const req: ListSmerdsRequest = {
        name: serviceName,
        limit: 10,
        environment: useEnvironmentStore.getState().selectedEnvironment,
    }
    return VelezAPI.ListSmerds(req, GetInitReq())
}

export async function FetchSmerd(name: string): Promise<ProtoSmerd> {
    const req: ListSmerdsRequest = {name, limit: 1, environment: useEnvironmentStore.getState().selectedEnvironment}
    return VelezAPI.ListSmerds(req, GetInitReq()).then((res) => {
        if (!res.smerds || res.smerds.length === 0) {
            throw new Error("Smerd not found")
        }
        return res.smerds[0]
    })
}


export async function GetSmerd(name: string, initReq: InitReq): Promise<Smerd> {
    const req = {
        name: name,
        limit: 1,
        environment: useEnvironmentStore.getState().selectedEnvironment,
    } as ListSmerdsRequest

    return VelezAPI.ListSmerds(req, initReq).then(
        (res) => {
            if (!res.smerds || res.smerds.length === 0) {
                throw new Error("Smerd not found")
            }

            return {
                name: res.smerds[0].name,
                imageName: res.smerds[0].imageName,
                ports: (res.smerds[0].ports || [])
                    .map((v) => {
                        return {
                            servicePort: v.servicePortNumber,
                            exposedPort: v.exposedTo,
                        } as Port
                    }),
                volumes: (res.smerds[0].volumes || [])
                    .map((v) => {
                        return {
                            containerPath: v.containerPath,
                            virtualVolume: v.volumeName,
                        } as Volume
                    })
            } as Smerd
        }
    )
}


export async function FetchNodeHardware(initReq: InitReq): Promise<GetHardwareResponse> {
    return VelezAPI.GetHardware({}, initReq)
}


export async function ListImages(
    name: string, initReq: InitReq, registryId?: string
): Promise<SearchImagesResponse> {
    const req: SearchImagesRequest = {
        name: name,
        registryId: registryId,
    }

    return VelezAPI.SearchImages(req, initReq)
}


// DeploySmerdStream is a pilot streaming counterpart to DeploySmerd: it
// enqueues the same create_smerd task (dedup'd on the smerd's name, same as
// the unary call) and forwards live TaskStatus updates to onStatus as the
// task progresses, instead of blocking until it's done. TaskStatus doesn't
// carry the created container's details, so once the stream ends the caller
// is expected to fetch the final Smerd separately (e.g. via GetSmerd).
export async function DeploySmerdStream(
    smerd: CreateSmerdReq,
    initReq: InitReq,
    onStatus: (status: TaskStatus) => void
): Promise<void> {
    return TasksApi.CreateSmerdStream(toProto(smerd), onStatus, initReq)
}


export async function DeploySmerd(smerd: CreateSmerdReq, initReq: InitReq): Promise<Smerd> {
    return VelezAPI.CreateSmerd(toProto(smerd), initReq).then((res) => {
        return {
            name: res.name,
            imageName: res.imageName,
            ports: (res.ports || [])
                .map((v) => {
                    return {
                        servicePort: v.servicePortNumber,
                        exposedPort: v.exposedTo,
                    } as Port
                }),
            volumes: (res.volumes || [])
                .map((v) => {
                    return {
                        containerPath: v.containerPath,
                        virtualVolume: v.volumeName,
                    } as Volume
                })
        } as Smerd
    })
}
