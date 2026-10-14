import {useEffect, useRef, useState} from "react"
import {ConfirmDialog, Dropdown, Toggle} from "@vervstack/chures"

import cls from "@/pages/auto-update/components/AutoUpdateRow/AutoUpdateRow.module.css"
import type {AutoUpdateBumpLevel, ListAutoUpdateServicesResponseService} from "@/app/api/velez/auto_update_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useTaskWatch} from "@/app/hooks/taskWatch/TaskWatch.ts"
import {useWatchedTask} from "@/app/hooks/taskWatch/WatchedTask.ts"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {queryClient} from "@/app/queryClient.ts"
import {
    AUTO_UPDATE_ACTION,
    BUMP_LEVEL_OPTIONS,
    bumpLevelName,
    confirmTitle,
    initialBumpLevel,
    isTaskFailed,
    restartWarning,
} from "@/processes/autoUpdate.ts"
import {isTaskInFlight, isTaskTerminal} from "@/processes/buildkitControl.ts"
import {AUTO_UPDATE_QUERY_KEY, SetServiceAutoUpdateMutation} from "@/processes/queries/autoUpdate.ts"

interface Props {
    service: ListAutoUpdateServicesResponseService
}

const DROPDOWN_OPTIONS = BUMP_LEVEL_OPTIONS.map((o) => ({id: o.id, name: o.name, title: o.hint}))

export default function AutoUpdateRow({service}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const watch = useTaskWatch((state) => state.watch)
    const setAutoUpdate = SetServiceAutoUpdateMutation()

    const name = service.name ?? ""
    const isEnabled = service.isEnabled ?? false
    const task = useWatchedTask(name, AUTO_UPDATE_ACTION)
    const isLocked = task !== undefined && !isTaskTerminal(task)
    const wasInFlight = useRef(false)
    const [level, setLevel] = useState(initialBumpLevel(service.bumpLevel))

    useEffect(() => {
        setLevel(initialBumpLevel(service.bumpLevel))
    }, [service.bumpLevel])

    useEffect(() => {
        if (isTaskInFlight(task)) {
            wasInFlight.current = true
            return
        }
        if (!wasInFlight.current || !isTaskTerminal(task)) return

        wasInFlight.current = false
        queryClient.invalidateQueries({queryKey: AUTO_UPDATE_QUERY_KEY})
        if (isTaskFailed(task)) {
            toaster.bake({title: "Error", description: `Auto-update change for ${name} failed.`, level: "Error"})
        }
    }, [task?.status])

    function openConfirm(nextEnabled: boolean, nextLevel: AutoUpdateBumpLevel) {
        function handleAccepted() {
            CloseDialog()
            setLevel(nextLevel)
            watch(name, AUTO_UPDATE_ACTION)
        }

        function handleConfirm() {
            return setAutoUpdate.mutateAsync({name, isEnabled: nextEnabled, bumpLevel: nextLevel})
                .then(handleAccepted)
                .catch(toaster.catchGrpc)
        }

        OpenDialog(
            <ConfirmDialog
                title={confirmTitle(isEnabled, nextEnabled)}
                message={restartWarning(name, nextEnabled)}
                confirmLabel="Confirm"
                onConfirm={handleConfirm}
                onClose={CloseDialog}
            />
        )
    }

    function handleToggle(checked: boolean) {
        openConfirm(checked, level)
    }

    function handleLevelChange(value: string[]) {
        const nextLevel = value[0] as AutoUpdateBumpLevel | undefined
        if (!nextLevel || nextLevel === level) return

        if (isEnabled) {
            openConfirm(true, nextLevel)
            return
        }
        setLevel(nextLevel)
    }

    function renderLevelControl() {
        if (isLocked) return <span className={cls.LockedLevel}>{bumpLevelName(level)}</span>
        return <Dropdown portal options={DROPDOWN_OPTIONS} value={[level]} onChange={handleLevelChange}/>
    }

    return (
        <div className={cls.AutoUpdateRowContainer}>
            <div className={cls.Row}>
                <span className={cls.Name}>{name}</span>
                <span className={cls.Cell}>{service.image || "-"}</span>
                <div className={cls.LevelWrapper}>{renderLevelControl()}</div>
                <div className={cls.EnabledWrapper}>
                    <Toggle checked={isEnabled} onChange={handleToggle} disabled={isLocked}/>
                    {isLocked && <span className={cls.Restarting}>Restarting…</span>}
                </div>
            </div>
        </div>
    )
}
