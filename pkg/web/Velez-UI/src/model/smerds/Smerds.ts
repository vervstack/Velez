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
