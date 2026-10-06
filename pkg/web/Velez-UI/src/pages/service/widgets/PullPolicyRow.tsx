import {Dropdown} from "@vervstack/chures"

import cls from "@/pages/service/widgets/PullPolicyRow.module.css"
import Button from "@/components/base/Button.tsx"
import {RunnerPullPolicy} from "@/app/api/velez"
import {availablePullPolicies} from "@/processes/runnerSettings.ts"

interface Props {
    index: number
    chain: RunnerPullPolicy[]
    onSelect: (index: number, policy: RunnerPullPolicy) => void
    onRemove: (index: number) => void
}

export default function PullPolicyRow({index, chain, onSelect, onRemove}: Props) {
    const options = availablePullPolicies(chain, index)

    function handleChange(ids: string[]) {
        const selected = options.find((option) => option.id === ids[0])
        if (!selected) return

        onSelect(index, selected.id)
    }

    function handleRemove() {
        onRemove(index)
    }

    return (
        <div className={cls.PullPolicyRowContainer}>
            <div className={cls.DropdownWrapper}>
                <Dropdown
                    label={index === 0 ? "Policy" : `Fallback ${index}`}
                    options={options}
                    value={[chain[index]]}
                    onChange={handleChange}
                    portal
                />
            </div>
            <Button sm variant="ghost" onClick={handleRemove}>✕</Button>
        </div>
    )
}
