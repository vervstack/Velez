import {Tooltip} from "react-tooltip";

import cls from "@/components/smerd/PortMapping.module.css";

import ArrowForward from "@/assets/icons/ArrowForward.svg";
import {getLinkToPort} from "@/model/services/Services.tsx";


interface PortMappingProps {
    portFrom: number;
    portTo: number
}

export default function PortMapping({portFrom, portTo}: PortMappingProps) {
    return (
        <div className={cls.PortMappingContainer}>
            <div className={cls.PortsPair}>
                <div
                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Service port"
                    data-tooltip-place="top"
                    className={cls.Port}>{portFrom}</div>
                <img
                    className={cls.MapArrow}
                    src={ArrowForward}
                    alt={'->'}
                />
                <div
                    className={cls.Port}
                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Exposed to port"
                    data-tooltip-place="bottom"
                >{portTo}</div>
            </div>
            <div className={cls.OpenPort}>
                <span
                    className="material-symbols-outlined"
                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Open port"
                    data-tooltip-place="top-start"
                    children={"open_in_new"}
                    onClick={()=> {window.open(getLinkToPort(portTo), "_blank")}}
                />
            </div>

            <Tooltip
                id={"open-port"}
            />
        </div>
    )
}
