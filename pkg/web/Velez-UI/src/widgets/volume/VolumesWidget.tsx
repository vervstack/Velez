import cls from '@/widgets/volume/VolumesWidget.module.css'

import {Volume} from "@/model/smerds/Smerds.ts";
import VolumeMapping from "@/components/smerd/VolumeMapping.tsx";

interface VolumesWidgetProps {
    volumes: Volume[]
}

export default function VolumesWidget({volumes}: VolumesWidgetProps) {
    return (
        <div className={cls.VolumesWidgetContainer}>
            <div>Volumes</div>
            <div className={cls.VolumeMappingWrapper}>
                {volumes.length != 0 ? volumes.map(v =>
                    <VolumeMapping
                        containerPath={v.containerPath}
                        volumeName={v.virtualVolume}/>
                ) : <div>No volumes</div>
                }
            </div>

        </div>
    )
}
