import {ConfirmDialog} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/RestartConfirm/RestartConfirm.module.css"
import {restartMessage} from "@/dialogs/CreateServiceDialog/processes/restartMessage.ts"

interface Props {
    containerName?: string
    hasPublishedPorts: boolean
    onConfirm(): void
    onCancel(): void
}

export default function RestartConfirm({containerName, hasPublishedPorts, onConfirm, onCancel}: Props) {
    return (
        <div className={cls.RestartConfirmContainer}>
            <ConfirmDialog
                danger
                title="Restart container?"
                message={restartMessage(containerName, hasPublishedPorts)}
                confirmLabel="Onboard and restart"
                onConfirm={onConfirm}
                onClose={onCancel}
            />
        </div>
    )
}
