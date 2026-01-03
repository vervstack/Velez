function keyPath() {
    return "velez_api_key"
}

function pathPrefixPath() {
    return "velez_path_prefix"
}

export function StoreApiKey(apiKey: string) {
    localStorage.setItem(keyPath(), apiKey)
}

export function StorePathPrefix(pathPrefix: string) {
    localStorage.setItem(pathPrefixPath(), pathPrefix)
}

export function GetInitReq(): InitReq {
    const key = localStorage.getItem(keyPath())
    if (!key) throw new Error("unauthorized: no apit key in settings");

    return {
        pathPrefix: localStorage.getItem(pathPrefixPath()) || "",
        headers: {
            'Grpc-Metadata-Authorization': key
        }
    }
}

export interface InitReq {
    pathPrefix: string
    headers: {
        'Grpc-Metadata-Authorization': string
    }
}
