import cls from '@/widgets/volume/VolumesWidget.module.css';
import PlusButtonSVG from "@/assets/icons/Plus.svg";

import {Volume} from "@/model/smerds/Smerds.ts";
import VolumeMapping from "@/components/smerd/VolumeMapping.tsx";

interface VolumesWidgetProps {
    volumes: Volume[],
    onChange?: (v: Volume[]) => void
}

export default function VolumesWidget({volumes, onChange}: VolumesWidgetProps) {

    function addVolume() {
        if (!onChange) return

        onChange(volumes)
        volumes.push({containerPath: '', virtualVolume: ''})
    }

    function onIndexUpdate(i: number) {
        if (!onChange) return undefined

        return (newContainerPath: string, newVolumeName: string) => {
            volumes[i].containerPath = newContainerPath
            volumes[i].virtualVolume = newVolumeName
            onChange(volumes)
        }
    }

    return (
        <div className={cls.VolumesWidgetContainer}>
            <div className={cls.HeaderWrapper}>
                <div>Volumes</div>
                {onChange ?
                    <div
                        className={cls.AddButton}
                        onClick={addVolume}
                    >
                        <img src={PlusButtonSVG} alt={'+'}/>
                    </div> : null}
            </div>
            <div className={cls.VolumeMappingWrapper}>
                {volumes.length != 0 ? volumes.map((v, i) =>
                    <div
                        key={i}
                    >
                        <VolumeMapping
                            containerPath={v.containerPath}
                            volumeName={v.virtualVolume}
                            onChange={onIndexUpdate(i)}
                        />
                    </div>
                ) : null
                }
            </div>

        </div>
    )
}
