import {ServiceType} from "@vervstack/velez"

import MatreshkaIcon from "@/assets/icons/matreshka/MatreshkaIcon";
import MakoshIcon from "@/assets/icons/makosh/MakoshIcon";
import UnknownServiceIcon from "@/assets/icons/UnknownServiceIcon.tsx";

export class Service {
    title: string
    icon: React.JSX.Element
    webLink?: string

    constructor(type: ServiceType, port?: number) {
        const serviceMeta = metaByType.get(type);
        if (!serviceMeta) {
            this.title = type.toString()
            this.icon = <UnknownServiceIcon/>
            return
        }

        this.title = serviceMeta.title;
        this.icon = serviceMeta.icon;

        if (port) {
            const {protocol, hostname} = window.location;
            this.webLink = `${protocol}//${hostname}:${port}`;
        }
    }
}

interface ServiceMeta {
    title: string
    icon: React.JSX.Element
}

const metaByType = new Map<ServiceType, ServiceMeta>();
metaByType.set(ServiceType.makosh, {title: "Makosh", icon: <MakoshIcon/>});
metaByType.set(ServiceType.matreshka, {title: "Matreshka", icon: <MatreshkaIcon/>});
metaByType.set(ServiceType.webserver, {title: "Angie (WebServer)", icon: <MatreshkaIcon/>});


