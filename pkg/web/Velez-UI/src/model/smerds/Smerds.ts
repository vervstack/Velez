import {
    CreateSmerdRequest,
    Volume as ProtoVolume,
    Port as ProtoPort,
} from "@vervstack/velez";

export interface Smerd {
    name: string
    imageName: string

    ports: Port[]
    volumes: Volume[]
}

export interface Port {
    servicePort: number
    exposedPort: number | null
}

export interface Volume {
    containerPath: string
    virtualVolume: string
}

export class CreateSmerdReq {
    name: string = ''
    imageName: string = ''
    command: string | null = null

    ignoreConfig: boolean = false
    autoUpgrade: boolean = false
    useImagePorts: boolean = false

    env: Record<string, string> = {}
    labels: Record<string, string> = {}

    ports: Port[] = []
    volumes: Volume[] = []
    binds: Volume[] = []

    constructor() {
    }
}

export function fromProto(proto: CreateSmerdRequest | undefined): CreateSmerdReq {
    const smerdReq = new CreateSmerdReq()

    if (!proto) {
        return smerdReq
    }

    smerdReq.name = proto.name || smerdReq.name
    smerdReq.imageName = proto.imageName || smerdReq.imageName
    smerdReq.command = proto.command || smerdReq.command

    smerdReq.ignoreConfig = proto.ignoreConfig || smerdReq.ignoreConfig
    smerdReq.autoUpgrade = proto.autoUpgrade || smerdReq.autoUpgrade
    smerdReq.useImagePorts = proto.useImagePorts || smerdReq.useImagePorts

    smerdReq.env = proto.env || smerdReq.env
    smerdReq.labels = proto.labels || smerdReq.labels

    smerdReq.volumes = fromProtoVolumes(proto.settings?.volumes) || smerdReq.volumes
    smerdReq.ports = fromProtoPorts(proto.settings?.ports) || smerdReq.ports
    smerdReq.binds = proto.settings?.binds?.map((v) => {
        if (!v.containerPath) return undefined

        return {
            containerPath: v.containerPath,
            virtualVolume: v.hostPath,
        } as Volume
    }).filter((v) => v != undefined) || smerdReq.binds
    return smerdReq
}

export function toProto(r: CreateSmerdReq): CreateSmerdRequest {
    return {
        name: r.name,
        imageName: r.imageName,
        command: r.command,

        ignoreConfig: r.ignoreConfig,
        useImagePorts: r.useImagePorts,
        autoUpgrade: r.autoUpgrade,

        env: r.env,
        labels: r.labels,

        hardware: undefined,
        settings: {
            ports: r.ports.map(p => {
                return {
                    servicePortNumber: p.servicePort,
                    exposedTo: p.exposedPort,
                }
            }),
            network: undefined,
            volumes: r.volumes.map(p=>{
                return {
                    containerPath: p.containerPath,
                    volumeName: p.virtualVolume,
                }
            }),
            binds: r.binds.map(p => {
                return {
                    containerPath: p.containerPath,
                    hostPath: p.virtualVolume,
                }
            }),
        },
        healthcheck: undefined,
        restart: undefined,
        config: undefined,
    } as CreateSmerdRequest
}


function fromProtoVolumes(v: ProtoVolume[] | undefined): Volume[] {
    if (!v) return []

    return v
        .map(fromProtoVolume)
        .filter(v => v != undefined)
}


function fromProtoVolume(v: ProtoVolume): Volume | undefined {
    if (!v.containerPath || !v.volumeName) return

    return {
        containerPath: v.containerPath,
        virtualVolume: v.volumeName,
    }
}


function fromProtoPorts(ports: ProtoPort[] | undefined): Port[] {
    if (!ports) return [];

    return ports
        .map(fromProtoPort)
        .filter((port): port is Port => port !== undefined);
}

function fromProtoPort(port: ProtoPort): Port | undefined {
    if (port.servicePortNumber === undefined) return undefined;

    return {
        servicePort: port.servicePortNumber,
        exposedPort: port.exposedTo || null,
    };
}
