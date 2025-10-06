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
            {
                ports.map(((v, i) => {
                    return (
                        <div key={i}>{
                            v.exposedPort ?
                                <PortMapping
                                    portFrom={v.servicePort}
                                    portTo={v.exposedPort}/>
                                :
                                null
                        }</div>
                    )
                }))
            }
        </div>
    )
}
