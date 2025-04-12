import {hookstate} from "@hookstate/core";

export type Settings = {
    apiURL: string
}

export const settingsHookState = hookstate(
    {
        apiURL: getBackendUrl()
    } as Settings)


function backendUrlLocalStorageKey(): string {
    return "backend_api_url"
}

export function getBackendUrl(): string {
    // let item = localStorage.getItem(backendUrlLocalStorageKey())
    const beApiAddr: string = import.meta.env.VITE_VELEZ_BACKEND_URL
    localStorage.setItem(backendUrlLocalStorageKey(), beApiAddr)
    return beApiAddr
}
