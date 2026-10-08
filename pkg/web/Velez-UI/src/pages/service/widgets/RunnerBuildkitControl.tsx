import {useEffect} from "react"

import cls from "@/pages/service/widgets/RunnerBuildkitControl.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useTaskWatch} from "@/app/hooks/taskWatch/TaskWatch.ts"
import {useWatchedTask} from "@/app/hooks/taskWatch/WatchedTask.ts"
import {queryClient} from "@/app/queryClient.ts"
import {
    buildkitControl,
    BUILDKIT_UNSUPPORTED_TOOLTIP,
    isTaskTerminal,
    SET_RUNNER_BUILDKIT_ACTION,
} from "@/processes/buildkitControl.ts"
import {RUNNERS_QUERY_KEY, SetRunnerBuildkitMutation, useListRunnersQuery} from "@/processes/queries/runners.ts"
import Button from "@/components/base/Button.tsx"
import InfoMark from "@/components/base/InfoMark.tsx"

interface Props {
    runnerName: string
}

export default function RunnerBuildkitControl({runnerName}: Props) {
    const toaster = useToaster()
    const watch = useTaskWatch((state) => state.watch)
    const task = useWatchedTask(runnerName, SET_RUNNER_BUILDKIT_ACTION)
    const runnersQuery = useListRunnersQuery()
    const setBuildkit = SetRunnerBuildkitMutation()

    const runner = runnersQuery.data?.runners?.find((r) => r.name === runnerName)
    const isEnabled = runner?.isBuildkitEnabled ?? false
    const isSupported = Boolean(runner?.dindName)
    const control = buildkitControl(task, isEnabled)

    useEffect(() => {
        if (isTaskTerminal(task)) {
            queryClient.invalidateQueries({queryKey: RUNNERS_QUERY_KEY})
        }
    }, [task?.status])

    function handleToggle() {
        setBuildkit.mutateAsync({name: runnerName, isBuildkitEnabled: !isEnabled})
            .then(() => watch(runnerName, SET_RUNNER_BUILDKIT_ACTION))
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.RunnerBuildkitControlContainer}>
            <Button
                onClick={handleToggle}
                disabled={!isSupported || control.isLocked || setBuildkit.isPending || runnersQuery.isLoading}
            >
                {control.label}
            </Button>

            {runnersQuery.isSuccess && !isSupported && <InfoMark tooltip={BUILDKIT_UNSUPPORTED_TOOLTIP}/>}
        </div>
    )
}
