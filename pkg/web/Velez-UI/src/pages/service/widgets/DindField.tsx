import {Dropdown, DropdownOption} from "@vervstack/chures"

import cls from "@/pages/service/widgets/DindField.module.css"

interface Props {
    value: string
    dindNames: string[]
    isLoading: boolean
    isBuildkitEnabled: boolean
    onChange: (value: string) => void
}

const HOST_SOCKET_ID = "__host_socket__"
const HOST_SOCKET_OPTION: DropdownOption = {id: HOST_SOCKET_ID, name: "Host socket (none)"}

export default function DindField({value, dindNames, isLoading, isBuildkitEnabled, onChange}: Props) {
    const options: DropdownOption[] = [
        HOST_SOCKET_OPTION,
        ...dindNames.map(function toOption(name) {
            return {id: name, name}
        }),
    ]

    function handleChange(ids: string[]) {
        const id = ids[0]
        if (id === undefined) return

        onChange(id === HOST_SOCKET_ID ? "" : id)
    }

    return (
        <div className={cls.DindFieldContainer}>
            <div className={cls.FieldWrapper}>
                <label className={cls.FieldLabel}>Docker daemon</label>
                <div className={cls.DropdownWrapper}>
                    <Dropdown
                        label="Docker daemon"
                        options={options}
                        value={[value || HOST_SOCKET_ID]}
                        onChange={handleChange}
                        isLoading={isLoading}
                        portal
                    />
                </div>
            </div>
            {isBuildkitEnabled && (
                <span className={cls.Hint}>BuildKit moves with the DinD after redeploy.</span>
            )}
        </div>
    )
}
