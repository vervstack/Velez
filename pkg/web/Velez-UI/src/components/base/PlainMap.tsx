import cls from "@/components/base/PlainMap.module.css"
import PlusSvg from "@/assets/icons/Plus.svg";
import MinusSvg from "@/assets/icons/Minus.svg";

import Input from "@/components/base/Input.tsx";

interface PlainMapProps {
    label?: string
    records: Record<string, string>
    onChange: (records: Record<string, string>) => void
}

export default function PlainMap({label, records, onChange}: PlainMapProps) {
    function onNameChange(oldName: string, newName: string) {
        if (newName == oldName) return

        records[newName] = records[oldName]
        delete records[oldName]
        onChange(records)
    }

    function onValueChange(fieldName: string): (newValue: string) => void {
        return (newValue) => {
            records[fieldName] = newValue
            onChange(records)
        }
    }

    function addNew() {
        records[''] = ''
        onChange(records)
    }

    function remove(fieldName: string) {
        delete records[fieldName]
        onChange(records)
    }

    return (
        <div className={cls.PlainMapContainer}>
            <div className={cls.HeaderLabel}>{label ? label : null}
                <div
                    onClick={addNew}
                    className={cls.ActionButton}
                >
                    <img src={PlusSvg} alt={'+'}/>
                </div>
            </div>
            <div className={cls.Entries}>
                {Object.entries(records)
                    .map(([k, v]) => {
                            return (
                                <div
                                    className={cls.Entry}
                                    key={k}>
                                    <KeyValue
                                        name={k}
                                        onNameChange={onNameChange}
                                        onValueChange={onValueChange(k)}
                                        value={v}
                                    />
                                    <div
                                        className={cls.ActionButton}
                                        onClick={() => remove(k)}
                                    >
                                        <img src={MinusSvg} alt={'-'}/>
                                    </div>
                                </div>
                            )
                        }
                    )}
            </div>
        </div>
    )
}

interface KeyValueProps {
    name: string,
    value: string,
    onNameChange: (oldName: string, newName: string) => void
    onValueChange: (newValue: string) => void
}

function KeyValue({name, value, onNameChange, onValueChange}: KeyValueProps) {
    function updateName(newName: string) {
        onNameChange(name, newName)
    }

    return (
        <div className={cls.KeyValueContainer}>
            <Input
                inputValue={name}
                onLeave={updateName}/>
            :
            <Input
                inputValue={value}
                onChange={(v) => onValueChange(v)}/>
        </div>
    )
}
