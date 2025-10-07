import {Tooltip} from "react-tooltip";

import cls from "@/components/smerd/VolumeMapping.module.css";

import ArrowForward from "@/assets/icons/ArrowForward.svg";

interface VolumeMappingProps {
    containerPath: string
    volumeName: string
}

export default function VolumeMapping({containerPath, volumeName}: VolumeMappingProps) {
    return (
        <div>
            <div className={cls.VolumeMappingContainer}>
                <div
                    children={containerPath}
                    className={cls.Volume}

                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Container path"
                    data-tooltip-place="top"
                />
                <img
                    className={cls.MapArrow}
                    src={ArrowForward}
                    alt={'->'}
                />
                <div
                    children={volumeName}

                    className={cls.Port}
                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Host path"
                    data-tooltip-place="bottom"
                />
            </div>

            <Tooltip
                id={"open-port"}
            />
        </div>
    )
}
