import {VervServiceType} from "@vervstack/velez"

import MatreshkaIcon from "@/assets/icons/services/matreshka.png";
import MakoshIcon from "@/assets/icons/services/makosh.png";
import PortainerIcon from "@/assets/icons/services/portainer.svg";
import HeadscaleIcon from "@/assets/icons/services/headscale.svg";
import AngieIcon from "@/assets/icons/services/angie.png";
import DatabasePixelIcon from "@/assets/icons/services/database-pixel.svg";

import UnknownServiceIcon from "@/assets/icons/unknown.svg";

export class Service {
    title: string
    icon: string
    webLink?: string
    description: string
    type: VervServiceType

    constructor(type: VervServiceType, port?: number) {
        const serviceMeta = metaByType.get(type);
        this.type = type;

        if (!serviceMeta) {
            this.title = type.toString()
            this.icon = UnknownServiceIcon
            this.description = `Unknown service. If only we knew what it is, but we don't know what it is`
            return
        }

        this.title = serviceMeta.title;
        this.icon = serviceMeta.icon;
        this.description = serviceMeta.description;

        if (port) {
            this.webLink = getLinkToPort(port)
        }
    }
}

interface ServiceMeta {
    title: string
    icon: string
    description: string
}

const metaByType = new Map<VervServiceType, ServiceMeta>();
metaByType.set(VervServiceType.makosh, {
    title: "Makosh",
    icon: MakoshIcon,
    description: `Verv's Standard service discovery inside cluster`
});
metaByType.set(VervServiceType.matreshka, {
    title: "Matreshka",
    icon: MatreshkaIcon,
    description: `Verv's Standard configuration system`
});
metaByType.set(VervServiceType.portainer, {
    title: "Portainer",
    icon: PortainerIcon,
    description: `Docker engine web interface`
});
metaByType.set(VervServiceType.headscale, {
    title: "Headscale",
    icon: HeadscaleIcon,
    description: `Verv's default private network manager.`
})

metaByType.set(VervServiceType.webserver, {
    title: "Angie (WebServer)",
    icon: AngieIcon,
    description: `Cluster entrypoint`
});

metaByType.set(VervServiceType.statefull_pg, {
    title: "Stateful mode",
    icon: DatabasePixelIcon,
    description: `
    Deploys a Postgres DB to enable stateful deployments and cluster joining. 
    The Verv API remains operational even if the state is unavailable, avoiding a single point of failure.`
});

export function getLinkToPort(port: number): string {
    const {protocol, hostname} = window.location;
    return `${protocol}//${hostname}:${port}`;
}

