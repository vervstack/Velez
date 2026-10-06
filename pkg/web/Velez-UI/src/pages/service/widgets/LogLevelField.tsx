import {Dropdown} from "@vervstack/chures"

import cls from "@/pages/service/widgets/LogLevelField.module.css"
import {RunnerLogLevel} from "@/app/api/velez"
import {LOG_LEVEL_OPTIONS} from "@/processes/runnerSettings.ts"

interface Props {
    value: RunnerLogLevel
    onChange: (value: RunnerLogLevel) => void
}

export default function LogLevelField({value, onChange}: Props) {
    function handleChange(ids: string[]) {
        const selected = LOG_LEVEL_OPTIONS.find((option) => option.id === ids[0])
        if (!selected) return

        onChange(selected.id)
    }

    return (
        <div className={cls.LogLevelFieldContainer}>
            <label className={cls.FieldLabel}>Log level</label>
            <div className={cls.DropdownWrapper}>
                <Dropdown
                    label="Level"
                    options={LOG_LEVEL_OPTIONS}
                    value={[value]}
                    onChange={handleChange}
                    portal
                />
            </div>
        </div>
    )
}
