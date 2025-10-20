import {Bind} from "@/model/smerds/Smerds.ts";

import ValuesMapping from "@/components/complex/mapping/ValuesMapping.tsx";
import ValuesPair from "@/components/complex/mapping/ValuesPair.tsx";

interface BindingsWidgetProps {
    bindings: Bind[],
    onChange?: (v: Bind[]) => void
}

export default function BindingWidget({bindings, onChange}: BindingsWidgetProps) {
    function addBind() {
        if (!onChange) return

        onChange(bindings)
        bindings.push({containerPath: '', hostPath: ''})
    }

    function onIndexUpdate(i: number) {
        if (!onChange) return undefined

        return (name: string, value: string | undefined) => {
            bindings[i].containerPath = name
            bindings[i].hostPath = value || ''
            onChange(bindings)
        }
    }

    function onIndexDelete(i: number) {
        if (!onChange) return undefined

        bindings = bindings.filter((_, index) => index != i)
        onChange(bindings)
    }

    return (
        <ValuesMapping
            onAdd={addBind}
            onDelete={onIndexDelete}
            header={'Bindings'}>
            {
                bindings.map((v, i) =>
                    <ValuesPair
                        name={v.containerPath}
                        value={v.hostPath}
                        onChange={onIndexUpdate(i)}
                    />
                )
            }
        </ValuesMapping>
    )
}
