import {
    VervPluginType,
    VervPluginState,
    Plugin as ApiPlugin
} from "@/app/api/velez";

import {

    Service,
} from "@/model/services/Services";

export function toServices(services: ApiPlugin[]): Service[] {
    const out: Service[] = []

    services.map(s => {
        const srv = new Service(s.type || VervPluginType.unknown_service_type, s.port)
        srv.state = s.state || VervPluginState.unknown
        out.push(srv)
    })

    return out
}
