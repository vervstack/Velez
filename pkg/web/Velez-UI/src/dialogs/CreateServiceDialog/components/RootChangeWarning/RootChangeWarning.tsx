import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/RootChangeWarning/RootChangeWarning.module.css"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import {rootSwitchMessage} from "@/dialogs/CreateServiceDialog/processes/rootSwitchMessage.ts"

interface Props {
    isVisible: boolean
    oldRootName: string
    newRootName: string
}

export default function RootChangeWarning({isVisible, oldRootName, newRootName}: Props) {
    const isClusterMode = IsStatefullModeEnabled()
    const message = rootSwitchMessage({oldRootName, newRootName, isClusterMode})

    return (
        <div
            role="status"
            aria-hidden={!isVisible}
            className={cn(cls.RootChangeWarningContainer, isVisible && cls.Visible)}
        >
            <div className={cls.Inner}>
                <strong className={cls.Lead}>Network root changes</strong>
                <span className={cls.Message}>{message}</span>
            </div>
        </div>
    )
}
