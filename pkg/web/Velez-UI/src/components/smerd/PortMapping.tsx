import {Tooltip} from "react-tooltip";
import OpenInTab from "@/assets/icons/OpenInTab.svg";

import cls from "@/components/smerd/PortMapping.module.css";

import ArrowForward from "@/assets/icons/ArrowForward.svg";
import {getLinkToPort} from "@/model/services/Services.tsx";


interface PortMappingProps {
    portFrom: number;
    portTo?: number
}

export default function PortMapping({portFrom, portTo}: PortMappingProps) {
    return (
        <div className={cls.PortMappingContainer}>
            <div className={cls.PortsPair}>
                <div
                    children={portFrom}
                    className={cls.Port}

                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Service port"
                    data-tooltip-place="top"
                    />
                <img
                    className={cls.MapArrow}
                    src={ArrowForward}
                    alt={'->'}
                />
                {portTo ?
                    <div
                        children={portTo}

                        className={cls.Port}
                        data-tooltip-id={"open-port"}
                        data-tooltip-content="Exposed to port"
                        data-tooltip-place="bottom"
                    /> : <div
                        children={'Not exposed'}
                    />
                }
            </div>
            {portTo ?
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
                :
                null
            }

            <Tooltip
                id={"open-port"}
            />
        </div>
    )
}
