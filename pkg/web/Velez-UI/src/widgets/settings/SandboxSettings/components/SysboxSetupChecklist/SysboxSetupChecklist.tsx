import {useMemo} from "react"

import Button from "@/components/base/Button.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"

import cls from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/SysboxSetupChecklist.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {
    RunSysboxSmokeTestMutation,
    useGetSettingsQuery,
    useSysboxStatusQuery,
} from "@/processes/queries/settings.ts"
import ChecklistStep from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/components/ChecklistStep/ChecklistStep.tsx"
import {
    deriveSysboxChecks,
    summarizeSysboxChecks,
} from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/processes/deriveSysboxChecks.ts"
import {
    SYSBOX_CHECKLIST_INTRO,
    SYSBOX_SETUP_STEPS,
} from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/processes/sysboxSetupSteps.ts"

export default function SysboxSetupChecklist() {
    const toaster = useToaster()
    const statusQuery = useSysboxStatusQuery()
    const settingsQuery = useGetSettingsQuery()
    const smokeTest = RunSysboxSmokeTestMutation()

    const isSysboxEnabled = settingsQuery.data?.settings?.isSysboxEnabled

    const checks = useMemo(
        () => deriveSysboxChecks(statusQuery.data, smokeTest.data, isSysboxEnabled),
        [statusQuery.data, smokeTest.data, isSysboxEnabled],
    )
    const summary = summarizeSysboxChecks(checks)

    function handleRecheck() {
        statusQuery.refetch()
    }

    function handleRunSmokeTest() {
        smokeTest.mutate(undefined, {onError: toaster.catchGrpc})
    }

    function renderSummary() {
        if (!statusQuery.data) {
            return null
        }
        return <span className={cls.Progress}>{summary.passed} of {summary.total} checks passed</span>
    }

    function renderSteps() {
        if (statusQuery.isLoading) {
            return (
                <div className={cls.Skeletons}>
                    {SYSBOX_SETUP_STEPS.map(step => <SkeletonLoader key={step.id} shape="block" width="100%" height="2.5rem"/>)}
                </div>
            )
        }
        if (statusQuery.isError) {
            return <QueryErrorState message="Failed to load Sysbox status." onRetry={handleRecheck}/>
        }
        return (
            <ol className={cls.Steps}>
                {SYSBOX_SETUP_STEPS.map((step, index) => (
                    <ChecklistStep key={step.id} step={step} check={checks[index]}/>
                ))}
            </ol>
        )
    }

    return (
        <details className={cls.SysboxSetupChecklistContainer}>
            <summary className={cls.Summary}>
                <span>Sysbox setup checklist (host requirements)</span>
                {renderSummary()}
            </summary>
            <p className={cls.Intro}>{SYSBOX_CHECKLIST_INTRO}</p>
            <div className={cls.Actions}>
                <Button sm onClick={handleRecheck} disabled={statusQuery.isFetching}>Re-check</Button>
                <Button sm onClick={handleRunSmokeTest} disabled={smokeTest.isPending}>
                    {smokeTest.isPending ? "Running…" : "Run smoke test"}
                </Button>
            </div>
            {renderSteps()}
        </details>
    )
}
