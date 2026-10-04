import cn from "classnames"

import cls from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/components/ChecklistStep/ChecklistStep.module.css"
import type {SysboxCheck, SysboxCheckState} from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/processes/deriveSysboxChecks.ts"
import type {SysboxSetupStep} from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/processes/sysboxSetupSteps.ts"

interface Props {
    step: SysboxSetupStep
    check: SysboxCheck
}

const STATE_GLYPH: Record<SysboxCheckState, string> = {
    pass: "✓",
    fail: "✗",
    unknown: "–",
}

const STATE_LABEL: Record<SysboxCheckState, string> = {
    pass: "Passed",
    fail: "Failed",
    unknown: "Not checked",
}

export default function ChecklistStep({step, check}: Props) {
    function renderDetail() {
        return (
            <>
                <p className={cls.Detail}>{step.detail}</p>
                {step.command && <pre className={cls.Code}>{step.command}</pre>}
                {step.expected && (
                    <p className={cls.Detail}>Expected: <code className={cls.InlineCode}>{step.expected}</code></p>
                )}
                {step.fallback && <pre className={cls.Code}>{step.fallback}</pre>}
            </>
        )
    }

    function renderBody() {
        if (check.state === "pass") {
            return (
                <details className={cls.Collapsed}>
                    <summary className={cls.DetailToggle}>Details</summary>
                    {renderDetail()}
                </details>
            )
        }
        return renderDetail()
    }

    return (
        <li className={cn(cls.ChecklistStepContainer, cls[check.state])}>
            <div className={cls.Header}>
                <span role="img" aria-label={STATE_LABEL[check.state]} className={cls.Indicator}>
                    {STATE_GLYPH[check.state]}
                </span>
                <span className={cls.Title}>{step.title}</span>
            </div>
            {check.hint && <p className={cls.Hint}>{check.hint}</p>}
            {renderBody()}
        </li>
    )
}
