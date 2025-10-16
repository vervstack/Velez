import {Volume} from "@/model/smerds/Smerds.ts";

import VolumeMapping from "@/components/smerd/VolumeMapping.tsx";
import ValuesMapping from "@/components/complex/mapping/ValuesMapping.tsx";

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

    function onIndexDelete(i: number) {
        if (!onChange) return undefined

        volumes = volumes.filter((_, index) => index != i)
        onChange(volumes)
    }

    return (
        <ValuesMapping
            onAdd={addVolume}
            onDelete={onIndexDelete}
            header={'Volumes'}>
            {
                volumes.map((v, i) =>
                    <VolumeMapping
                        containerPath={v.containerPath}
                        volumeName={v.virtualVolume}
                        onChange={onIndexUpdate(i)}
                    />
                )
            }
        </ValuesMapping>
    )
}
