import {Checkbox} from "@vervstack/chures"
import {RunnerPullPolicy} from "@/app/api/velez"

interface Props {
    policy: RunnerPullPolicy
    label: string
    isChecked: boolean
    onToggle: (policy: RunnerPullPolicy, isChecked: boolean) => void
}

export default function AllowedPolicyOption({policy, label, isChecked, onToggle}: Props) {
    function handleChange(checked: boolean) {
        onToggle(policy, checked)
    }

    return <Checkbox label={label} checked={isChecked} onChange={handleChange}/>
}
