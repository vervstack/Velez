import {create} from 'zustand';
import {InitReq, SettingsBase} from "@/app/settings/state.ts";

type Credentials = {
    token: string
    url: string

    setToken: (v: string) => void
    setUrl: (v: string) => void
    getInitReq: () => InitReq

}

const ls = getFromLocalStorage()

export const useCredentialsStore = create<Credentials>((set, get) => ({
    url: ls.backendUrl,
    token: ls.authHeader,

    setToken: (token: string) => {
        set({token})

        storeToLocalStorage({
            backendUrl: get().url,
            authHeader: token
        })
    },
    setUrl: (url: string) => {
        set({url})

        storeToLocalStorage({
            backendUrl: url,
            authHeader: get().token
        })
    },
    getInitReq: () => {
        const {url, token} = get();

        return {
            pathPrefix: url,
            headers: {
                'Grpc-Metadata-Authorization': token,
            }
        }
    }
}))

function storeToLocalStorage(value: SettingsBase) {
    localStorage.setItem("settings", JSON.stringify(value))
}

function getFromLocalStorage(): SettingsBase {
    return JSON.parse(localStorage.getItem("settings") || "{}")
}
