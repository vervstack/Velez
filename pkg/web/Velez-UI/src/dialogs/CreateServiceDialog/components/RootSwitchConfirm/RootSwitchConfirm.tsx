import {ConfirmDialog} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/RootSwitchConfirm/RootSwitchConfirm.module.css"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import {RootSwitchInfo, rootSwitchMessage} from "@/dialogs/CreateServiceDialog/processes/rootSwitchMessage.ts"

interface Props {
    oldRootName: string
    newRootName: string
    onConfirm(): void
    onCancel(): void
}

export default function RootSwitchConfirm({oldRootName, newRootName, onConfirm, onCancel}: Props) {
    const isClusterMode = IsStatefullModeEnabled()
    const info: RootSwitchInfo = {oldRootName, newRootName, isClusterMode}

    return (
        <div className={cls.RootSwitchConfirmContainer}>
            <ConfirmDialog
                danger
                title="Make network root?"
                message={rootSwitchMessage(info)}
                confirmLabel="Make network root"
                onConfirm={onConfirm}
                onClose={onCancel}
            />
        </div>
    )
}
