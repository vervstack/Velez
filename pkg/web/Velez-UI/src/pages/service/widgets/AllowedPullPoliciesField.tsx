import cls from "@/pages/service/widgets/AllowedPullPoliciesField.module.css"
import AllowedPolicyOption from "@/pages/service/widgets/AllowedPolicyOption.tsx"
import {RunnerPullPolicy} from "@/app/api/velez"
import {PULL_POLICY_OPTIONS, toggleAllowedPolicy} from "@/processes/runnerSettings.ts"

interface Props {
    value: RunnerPullPolicy[]
    onChange: (value: RunnerPullPolicy[]) => void
}

export default function AllowedPullPoliciesField({value, onChange}: Props) {
    function handleToggle(policy: RunnerPullPolicy, isChecked: boolean) {
        onChange(toggleAllowedPolicy(value, policy, isChecked))
    }

    return (
        <div className={cls.AllowedPullPoliciesFieldContainer}>
            <label className={cls.FieldLabel}>Allowed pull</label>
            <div className={cls.ChecksWrapper}>
                {PULL_POLICY_OPTIONS.map((option) => (
                    <AllowedPolicyOption
                        key={option.id}
                        policy={option.id}
                        label={option.name}
                        isChecked={value.includes(option.id)}
                        onToggle={handleToggle}
                    />
                ))}
                {value.length === 0 && <span className={cls.AnyHint}>Any (default)</span>}
            </div>
        </div>
    )
}
