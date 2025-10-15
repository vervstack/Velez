import {Tooltip} from "react-tooltip";

import cls from "@/components/smerd/VolumeMapping.module.css";

import ArrowForward from "@/assets/icons/ArrowForward.svg";
import Input from "@/components/base/Input.tsx";

interface VolumeMappingProps {
    containerPath: string
    volumeName: string
    onChange?: (containerPath: string, volumeName: string) => void
}

export default function VolumeMapping({containerPath, volumeName, onChange}: VolumeMappingProps) {

    const handleContainerPath = onChange ? (v: string) => onChange(v, volumeName) : undefined
    const handleVolumeName = onChange ? (v: string) => onChange(containerPath, v) : undefined

    return (
        <div className={cls.VolumeMappingContainer}>
            <div className={cls.VolumesPair}>
                <div
                    className={cls.Volume}

                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Container path"
                    data-tooltip-place="top"
                >
                    <Input
                        inputValue={containerPath}
                        onChange={handleContainerPath}
                    />
                </div>

                <img
                    className={cls.MapArrow}
                    src={ArrowForward}
                    alt={'->'}
                />

                <div
                    className={cls.Volume}
                    data-tooltip-id={"open-port"}
                    data-tooltip-content="Host path"
                    data-tooltip-place="bottom"
                >
                    <Input
                        inputValue={volumeName}
                        onChange={handleVolumeName}
                    />
                </div>
            </div>

            <Tooltip
                id={"open-port"}
            />
        </div>
    )
}
