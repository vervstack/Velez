import {
    VervServiceType,
    VervServiceState,
    VervService as ApiServices
} from "@vervstack/velez";

import {

    Service,
} from "@/model/services/Services";

export function toServices(services: ApiServices[]): Service[] {
    const out: Service[] = []

    services.map(s => {
        const srv = new Service(s.type || VervServiceType.unknown_service_type, s.port)
        srv.state = s.state || VervServiceState.unknown
        out.push(srv)
    })

    return out
}
