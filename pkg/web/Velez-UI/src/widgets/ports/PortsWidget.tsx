import cls from "@/widgets/ports/PortsWidget.module.css"

import {Port} from "@/model/smerds/Smerds.ts";
import PortMapping from "@/components/smerd/PortMapping.tsx";

interface PortsWidgetProps {
    ports: Port[]
}

export default function PortsWidget({ports}: PortsWidgetProps) {
    return (
        <div className={cls.PortWidgetContainer}>

            <div>Ports:</div>
            {ports.map(((v, i) =>
                    <div
                        className={cls.PortMappingWrapper}
                        key={i}>
                        <PortMapping
                            portFrom={v.servicePort}
                            portTo={v.exposedPort}/>
                    </div>
            ))}
        </div>
    )
}
