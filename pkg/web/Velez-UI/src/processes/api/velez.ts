import {ListSmerdsRequest, VelezAPI} from "@vervstack/velez";
import {InitReq} from "@/app/settings/state.ts";
import {Port, Smerd, Volume} from "@/model/smerds/Smerds.ts";

export async function ListSmerds(req: ListSmerdsRequest, initReq: InitReq) {
    req.limit = req.limit || 10
    return VelezAPI.ListSmerds(req, initReq)
}


export async function GetSmerd(name: string, initReq: InitReq): Promise<Smerd> {
    const req = {
        name: name,
        limit: 1,
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
