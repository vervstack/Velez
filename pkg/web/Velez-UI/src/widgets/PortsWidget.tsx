import {Port} from "@/model/smerds/Smerds.ts";

import PortMapping from "@/components/smerd/PortMapping.tsx";

import ValuesMapping from "@/components/complex/mapping/ValuesMapping.tsx";

interface PortsWidgetProps {
    ports: Port[]
    onChange?: (ports: Port[]) => void

    isRunning?: boolean
}

export default function PortsWidget({ports, onChange}: PortsWidgetProps) {
    function addPort() {
        if (!onChange) {
            return
        }

        ports.push({servicePort: 80, exposedPort: 80})
        onChange(ports)
    }

    function onIndexDelete(index: number) {
        if (!onChange) {
            return
        }

        ports = ports.filter((_, i) => i !== index)
        onChange(ports)
    }

    function onPortChange(index: number) {
        return (portFrom: number, portTo: number | null) => {
            if (!onChange) return undefined

            ports[index].servicePort = portFrom
            ports[index].exposedPort = portTo
            onChange(ports)
        }
    }

    return (
        <ValuesMapping
            header={"Ports:"}
            onAdd={addPort}
            onDelete={onIndexDelete}
        >
            {ports.map((p: Port, index: number) =>
                <PortMapping
                    portFrom={p.servicePort}
                    portTo={p.exposedPort}
                    onChange={onPortChange(index)}/>
            )}
        </ValuesMapping>
    )
}
