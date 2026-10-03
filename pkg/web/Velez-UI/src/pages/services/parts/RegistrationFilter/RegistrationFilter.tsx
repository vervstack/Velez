import {Dropdown} from "@vervstack/chures"
import type {DropdownOption} from "@vervstack/chures"

import type {RegistrationFilter as Filter} from "@/pages/services/processes/partitionByRegistration.ts"

const OPTIONS: DropdownOption[] = [
    {id: "all", name: "All"},
    {id: "registered", name: "Registered only"},
    {id: "unregistered", name: "Unregistered only"},
]

interface Props {
    value: Filter
    onChange: (value: Filter) => void
}

export default function RegistrationFilter({value, onChange}: Props) {
    function handleChange(selected: string[]) {
        if (selected.length === 0) return
        onChange(selected[0] as Filter)
    }

    return <Dropdown options={OPTIONS} value={[value]} onChange={handleChange} portal/>
}
