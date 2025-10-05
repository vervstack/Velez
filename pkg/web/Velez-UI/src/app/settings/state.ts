import {useEffect, useState} from "react";

interface SettingsBase {
    backendUrl: string
    authHeader: string;
}

interface Settings extends SettingsBase {
    setBackendUrl: (url: string) => void
    setAuthHeader: (h: string) => void;

    initReq: () => InitReq
}

export interface InitReq {
    pathPrefix: string
    headers: {
        'Grpc-Metadata-Authorization': string
    }
}

const settings = getFromLocalStorage("settings");

export default function useSettings(): Settings {
    const [backendUrl, setBackendUrl] = useState(settings.backendUrl)
    const [authHeader, setAuthHeader] = useState(settings.authHeader)


    useEffect(() => {
        storeToLocalStorage("settings", {backendUrl, authHeader})
    }, [backendUrl, authHeader]);

    function initReq() {
        return {
            pathPrefix: backendUrl,
            headers: {
                'Grpc-Metadata-Authorization': authHeader,
            }
        }
    }

    return {
        backendUrl,
        setBackendUrl,

        authHeader,
        setAuthHeader,

        initReq
    }
}

function storeToLocalStorage(key: string, value: SettingsBase) {
    localStorage.setItem(key, JSON.stringify(value))
}

function getFromLocalStorage(key: string): SettingsBase {
    return JSON.parse(localStorage.getItem(key) || "{}")
}
