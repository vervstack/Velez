import cls from "@/widgets/settings/AddressesSettings/components/RebuildProgress/RebuildProgress.module.css"
import type {AddressesRebuildStatus} from "@/model/settings/AddressesRebuildStatus.ts"

interface Props {
    status: AddressesRebuildStatus
}

export default function RebuildProgress({status}: Props) {
    if (status.isRunning) {
        return (
            <div className={cls.RebuildProgressContainer}>
                <div className={cls.Label}>
                    Rebuilding addresses… {status.doneSteps}/{status.totalSteps}
                </div>
                <progress className={cls.Bar} value={status.doneSteps} max={Math.max(status.totalSteps, 1)}/>
            </div>
        )
    }

    if (status.lastError !== "") {
        return (
            <div className={cls.RebuildProgressContainer}>
                <div className={cls.ErrorLine}>Last rebuild failed: {status.lastError}</div>
            </div>
        )
    }

    return null
}
