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
        out.push(new Service(s.type || ServiceType.unknown_service_type, s.port))
    })

    return out
}
