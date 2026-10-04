import {Toggle} from "@vervstack/chures"

import cls from "@/widgets/settings/SandboxSettings/components/SandboxToggleRow/SandboxToggleRow.module.css"

interface Props {
    label: string
    description: string
    isChecked: boolean
    isDisabled: boolean
    onChange(isChecked: boolean): void
}

export default function SandboxToggleRow({label, description, isChecked, isDisabled, onChange}: Props) {
    return (
        <div className={cls.SandboxToggleRowContainer}>
            <Toggle label={label} checked={isChecked} onChange={onChange} disabled={isDisabled}/>
            <p className={cls.Description}>{description}</p>
        </div>
    )
}
