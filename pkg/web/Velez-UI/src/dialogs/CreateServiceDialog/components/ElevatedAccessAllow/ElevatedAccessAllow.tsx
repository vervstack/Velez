import {Checkbox} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/ElevatedAccessAllow/ElevatedAccessAllow.module.css"

const EXPLANATION = "This image is not whitelisted. Allowing it lets this container keep its " +
    "privileged mode, devices and capabilities."

interface Props {
    repository: string
    isChecked: boolean
    onChange(isChecked: boolean): void
}

export default function ElevatedAccessAllow({repository, isChecked, onChange}: Props) {
    return (
        <div className={cls.ElevatedAccessAllowContainer}>
            <Checkbox
                label={`Allow elevated access for ${repository}`}
                checked={isChecked}
                onChange={onChange}
            />
            <span className={cls.Hint}>{EXPLANATION}</span>
        </div>
    )
}
