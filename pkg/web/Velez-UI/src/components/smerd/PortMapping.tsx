import {Tooltip} from "react-tooltip";
import OpenInTab from "@/assets/icons/OpenInTab.svg";

import cls from "@/components/smerd/PortMapping.module.css";

import ArrowForward from "@/assets/icons/ArrowForward.svg";
import {getLinkToPort} from "@/model/services/Services.tsx";
import Input from "@/components/base/Input.tsx";

interface PortMappingProps {
    portFrom: number;
    portTo: number | null;

    onChange: (portFrom: number, portTo: number | null) => void;

    isRunning?: boolean
}

export default function PortMapping({portFrom, portTo, onChange, isRunning}: PortMappingProps) {
    return (
        <div className={cls.PortMappingContainer}>
            <div className={cls.PortsPair}>
                <div
                    className={cls.Port}

                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Service port"
                    data-tooltip-place="top"
                >
                    <Input
                        inputValue={portFrom > 0 ? portFrom.toString() : ''}
                        onChange={(v: string) => {
                            if (v == '') {
                                onChange(0, portTo)
                            }
                            const newP = parseInt(v)
                            if (isNaN(newP)) return
                            onChange(parseInt(v), portTo)
                        }}
                    />
                    <img
                        className={cls.MapArrow}
                        src={ArrowForward}
                        alt={'->'}
                    />
                </div>
                <div
                    className={cls.Port}
                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Exposed to port"
                    data-tooltip-place="bottom"
                >
                    <Input
                        label={isRunning ? 'Not exposed' :
                            portTo ? undefined : 'Port will be exposed once you deploy'}
                        inputValue={(portTo || 0) > 0 ? portTo?.toString() : ''}
                        onChange={(v: string) => {
                            if (v == '') {
                                onChange(portFrom, null)
                            }

                            const newP = parseInt(v)
                            if (isNaN(newP)) return
                            onChange(portFrom, newP)
                        }}
                    />
                </div>
            </div>
            {portTo && isRunning ?
                <div className={cls.OpenPortButton}>
                    <img
                        src={OpenInTab}
                        alt={'^'}

                        className={cls.OpenPort}

                        data-tooltip-id={"open-port"}
                        data-tooltip-content="Open port"
                        data-tooltip-place="top-start"
                        onClick={() => {
                            window.open(getLinkToPort(portTo), "_blank")
                        }}
                    />
                </div>
                :
                null
            }

            <Tooltip
                id={"open-port"}
            />
        </div>
    )
}
