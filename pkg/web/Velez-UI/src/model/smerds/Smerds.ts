export interface Smerd {
    name: string
    imageName: string

    ports: Port[]
    volumes: Volume[]
}

export interface Port {
    servicePort: number
    exposedPort?: number
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


    ports: Port[] = []
    volumes: Volume[] = []
    env: Record<string, string> = {}
    labels: Record<string, string> = {}

    constructor() {}
}
