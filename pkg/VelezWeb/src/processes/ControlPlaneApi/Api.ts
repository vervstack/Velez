import {ControlPlane, ListServicesRequest} from "@godverv/velez"
import {getBackendUrl, settingsHookState} from "@/app/store/Settings";
import {toServices} from "@/processes/ControlPlaneApi/mappings/Services";

const initReq = {pathPrefix: getBackendUrl()};

export function setBackendUrl(url: string) {
    initReq.pathPrefix = url
}

export function ListServices() {
    const req: ListServicesRequest = {} as ListServicesRequest

    return ControlPlane.ListServices(req, initReq)
        .then(toServices)
}
