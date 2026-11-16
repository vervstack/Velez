import {useEffect, useMemo} from "react"

import {Port} from "@/model/smerds/Smerds.ts";

import PortMapping from "@/components/smerd/PortMapping.tsx";

import ValuesMapping from "@/components/complex/mapping/ValuesMapping.tsx";
import {expandAvailablePorts, isPortAvailable} from "@/processes/availablePorts.ts";
import {useVersionQuery} from "@/processes/queries/version.ts";

interface PortsWidgetProps {
    ports: Port[]
    onChange?: (ports: Port[]) => void

    isRunning?: boolean
    onValidityChange?: (hasInvalidPorts: boolean) => void
}

export default function PortsWidget({ports, onChange, onValidityChange}: PortsWidgetProps) {
    const versionQuery = useVersionQuery()
    const configuredRanges = versionQuery.data?.availablePorts ?? []
    const availablePorts = useMemo(() => expandAvailablePorts(configuredRanges), [configuredRanges])
    const isPortPoolKnown = versionQuery.data?.availablePorts !== undefined
    const hasInvalidPorts = isPortPoolKnown && ports.some((port) => {
        return port.exposedPort !== null && port.exposedPort > 0 && !isPortAvailable(availablePorts, port.exposedPort)
    })

    useEffect(() => {
        onValidityChange?.(hasInvalidPorts)
    }, [hasInvalidPorts, onValidityChange])

    function exposedPortError(port: Port): string | undefined {
        if (!isPortPoolKnown || port.exposedPort === null || port.exposedPort <= 0 ||
            isPortAvailable(availablePorts, port.exposedPort)) {
            return undefined
        }

        const ranges = configuredRanges.length > 0 ? configuredRanges.join(", ") : "no ports configured"

        return `Port ${port.exposedPort} is outside this node's allowed pool (ENVIRONMENT_AVAILABLE_PORTS: ${ranges}). ` +
            "Ask the node admin to add the port to the pool on the node and restart Velez."
    }

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
                    onChange={onPortChange(index)}
                    exposedPortError={exposedPortError(p)}/>
            )}
        </ValuesMapping>
    )
}
