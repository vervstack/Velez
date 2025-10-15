import cls from "@/widgets/ports/PortsWidget.module.css"
import PlusButtonSVG from "@/assets/icons/Plus.svg";

import {Port} from "@/model/smerds/Smerds.ts";
import PortMapping from "@/components/smerd/PortMapping.tsx";
import MinusSvg from "@/assets/icons/Minus.svg";

interface PortsWidgetProps {
    ports: Port[]
    onChange?: (ports: Port[]) => void

    isRunning?: boolean
}

export default function PortsWidget({ports, onChange, isRunning}: PortsWidgetProps) {

    function addPort() {
        if (!onChange) {
            return
        }

        ports.push({servicePort: 80, exposedPort: 80})
        onChange(ports)
    }

    function remove(index: number) {
        if (!onChange) {
            return
        }

        ports.splice(index, 1)
        onChange(ports)
    }

    return (
        <div className={cls.PortWidgetContainer}>
            <div className={cls.HeaderWrapper}>
                <div>Ports:</div>
                {onChange ?
                    <div
                        onClick={addPort}
                        className={cls.AddButton}>
                        <img src={PlusButtonSVG} alt={'+'}/>
                    </div> : null}
            </div>
            {ports.map(((v, i) =>
                    <div
                        className={cls.PortMappingWrapper}
                        key={i}>
                        <PortMapping
                            portFrom={v.servicePort}
                            portTo={v.exposedPort}
                            isRunning={isRunning}
                            onChange={(portFrom, portTo) => {
                                ports[i].servicePort = portFrom
                                ports[i].exposedPort = portTo

                                if (onChange) onChange(ports)
                            }}
                        />
                        <div
                            className={cls.ActionButton}
                            onClick={() => remove(i)}
                        >
                            <img src={MinusSvg} alt={'-'}/>
                        </div>
                    </div>
            ))}
        </div>
    )
}
