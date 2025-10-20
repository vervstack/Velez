import {Volume} from "@/model/smerds/Smerds.ts";

import ValuesMapping from "@/components/complex/mapping/ValuesMapping.tsx";
import ValuesPair from "@/components/complex/mapping/ValuesPair.tsx";

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

        return (newContainerPath: string, newVolumeName: string | undefined) => {
            volumes[i].containerPath = newContainerPath
            volumes[i].virtualVolume = newVolumeName || ""
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
                    <ValuesPair
                        name={v.containerPath}
                        value={v.virtualVolume}
                        onChange={onIndexUpdate(i)}
                    />
                )
            }
        </ValuesMapping>
    )
}
