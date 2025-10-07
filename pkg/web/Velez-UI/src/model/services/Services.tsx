import {ServiceType, CreateSmerdRequest} from "@vervstack/velez"

import MatreshkaIcon from "@/assets/icons/matreshka/MatreshkaIcon";
import MakoshIcon from "@/assets/icons/makosh/MakoshIcon";
import UnknownServiceIcon from "@/assets/icons/UnknownServiceIcon.tsx";

export class Service {
    title: string
    icon: React.JSX.Element
    webLink?: string
    description: string

    smerdConstructor?: CreateSmerdRequest

    constructor(type: ServiceType, port?: number) {
        const serviceMeta = metaByType.get(type);
        if (!serviceMeta) {
            this.title = type.toString()
            this.icon = <UnknownServiceIcon/>
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
    icon: React.JSX.Element
    description: string
}

const metaByType = new Map<ServiceType, ServiceMeta>();
metaByType.set(ServiceType.makosh, {
    title: "Makosh",
    icon: <MakoshIcon/>,
    description: `Verv Standard service discovery inside cluster`
});
metaByType.set(ServiceType.matreshka, {
    title: "Matreshka",
    icon: <MatreshkaIcon/>,
    description: `Verv Standard configuration system`
});
// metaByType.set(ServiceType.webserver, {
//     title: "Angie (WebServer)",
//     icon: <MatreshkaIcon/>,
//     description: ``
// });



export function getLinkToPort(port: number): string {
    const {protocol, hostname} = window.location;
    return `${protocol}//${hostname}:${port}`;
}

