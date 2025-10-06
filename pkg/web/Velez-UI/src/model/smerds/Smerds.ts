export interface Smerd {
    name: string
    imageName: string

    ports: Port[]
}

export interface Port {
    servicePort: number
    exposedPort?: number
}
