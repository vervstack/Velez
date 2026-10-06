import cls from "@/pages/service/widgets/PullPolicyField.module.css"
import PullPolicyRow from "@/pages/service/widgets/PullPolicyRow.tsx"
import Button from "@/components/base/Button.tsx"
import {RunnerPullPolicy} from "@/app/api/velez"
import {canAddFallback, nextFallbackPolicy} from "@/processes/runnerSettings.ts"

interface Props {
    value: RunnerPullPolicy[]
    onChange: (value: RunnerPullPolicy[]) => void
}

export default function PullPolicyField({value, onChange}: Props) {
    function handleAdd() {
        const next = nextFallbackPolicy(value)
        if (next === undefined) return

        onChange([...value, next])
    }

    function handleSelect(index: number, policy: RunnerPullPolicy) {
        onChange(value.map((current, i) => i === index ? policy : current))
    }

    function handleRemove(index: number) {
        onChange(value.filter((_, i) => i !== index))
    }

    const isEmpty = value.length === 0

    return (
        <div className={cls.PullPolicyFieldContainer}>
            <label className={cls.FieldLabel}>Pull policy</label>
            <div className={cls.ChainWrapper}>
                {value.map((policy, index) => (
                    <PullPolicyRow
                        key={policy}
                        index={index}
                        chain={value}
                        onSelect={handleSelect}
                        onRemove={handleRemove}
                    />
                ))}
                {isEmpty && <span className={cls.DefaultHint}>Default (always)</span>}
                {canAddFallback(value) && (
                    <div className={cls.AddRow}>
                        <Button sm onClick={handleAdd}>{isEmpty ? "Add" : "Add fallback"}</Button>
                    </div>
                )}
            </div>
        </div>
    )
}
