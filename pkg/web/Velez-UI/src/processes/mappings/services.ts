import {
    ServiceType,
    Service as ApiServices
} from "@vervstack/velez";

import {

    Service,
} from "@/model/services/Services";

export function toServices(services: ApiServices[]): Service[] {
    const out: Service[] = []

    services.map(s => {
        const srv = new Service(s.type || ServiceType.unknown_service_type, s.port)

        if (s.constructor !== Object)
            srv.smerdConstructor = s.constructor
        out.push(srv)
    })

    return out
}
