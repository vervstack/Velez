import {useState} from "react"

import cls from "@/dialogs/EnablePgReplicationDialog/EnablePgReplicationDialog.module.css"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import {EnablePgReplicationMutation} from "@/processes/queries/pg_instances.ts"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

interface Props {
    name: string
}

const SETTINGS = [
    "wal_level = replica (logical also meets this requirement)",
    "wal_log_hints = on",
    "max_wal_senders >= 10",
    "max_replication_slots >= 10",
]

export default function EnablePgReplicationDialog({name}: Props) {
    const {CloseDialog, LockClosing, UnlockClosing} = useDialog()
    const toaster = useToaster()
    const enablePgReplication = EnablePgReplicationMutation()
    const [isConfirmed, setIsConfirmed] = useState(false)

    function handleConfirm() {
        setIsConfirmed(true)
    }

    function handleStart() {
        LockClosing()
        return enablePgReplication.mutateAsync({name}).finally(UnlockClosing)
    }

    function handleSuccess() {
        toaster.bake({
            title: "Replication settings applied",
            description: `${name} is ready for physical replication.`,
            level: "Info",
        })
        CloseDialog()
    }

    function renderSetting(setting: string) {
        return <li key={setting}>{setting}</li>
    }

    if (isConfirmed) {
        return <div className={cls.EnablePgReplicationDialogContainer}>
            <DialogShell title="Enabling replication" onClose={CloseDialog} isFlush>
                <TaskProgressScreen
                    title="Enabling replication"
                    metaLine={name}
                    start={handleStart}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                />
            </DialogShell>
        </div>
    }

    return <div className={cls.EnablePgReplicationDialogContainer}>
        <DialogShell title="Enable replication" onClose={CloseDialog}>
            <ul className={cls.Settings}>{SETTINGS.map(renderSetting)}</ul>
            <p className={cls.Warning}>The server will be restarted to apply these settings.</p>
            <div className={cls.ActionsWrapper}>
                <Button variant="secondary" onClick={CloseDialog}>No</Button>
                <Button variant="primary" onClick={handleConfirm}>Yes</Button>
            </div>
        </DialogShell>
    </div>
}
