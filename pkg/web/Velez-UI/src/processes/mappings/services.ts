import {
    VervPluginType,
    VervPluginState,
    Plugin as ApiPlugin
} from "@/app/api/velez";

import {
    VervPlugin,
} from "@/model/services/VervPlugins.tsx";

export function toServices(services: ApiPlugin[]): VervPlugin[] {
    const out: VervPlugin[] = []

    services.map(s => {
        const srv = new VervPlugin(s.type || VervPluginType.unknown_service_type, s.serviceName || "Unspecified")
        srv.state = s.state || VervPluginState.unknown
        out.push(srv)
    })

    return out
}
