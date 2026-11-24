import {ConfirmDialog} from "@vervstack/chures"
import cn from "classnames"
import {Link} from "react-router-dom"

import type {HaproxyInstance} from "@/app/api/velez/haproxy_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {Routes} from "@/app/router/Routes.ts"
import Button from "@/components/base/Button.tsx"
import StatusDot from "@/components/base/StatusDot.tsx"
import cls from "@/pages/haproxy/HaproxyPage.module.css"
import {useDropHaproxyInstanceMutation} from "@/processes/queries/haproxy.ts"

// Mirrors labels.HaproxyNamePrefix in internal/domain/labels/verv_labels.go.
const HAPROXY_NAME_PREFIX = "haproxy_"

interface Props {
    instance: HaproxyInstance
    isSelected: boolean
    onSelect(name: string): void
}

export default function HaproxyInstanceRow({instance, isSelected, onSelect}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const dropInstance = useDropHaproxyInstanceMutation()
    const name = instance.name ?? ""
    const status = instance.status ?? ""
    const dotStatus = status === "running" ? "running" : status === "restarting" ? "degraded" : "stopped"

    function handleSelect() {
        onSelect(name)
    }

    function handleConfirmDrop() {
        return dropInstance.mutateAsync(name).then(CloseDialog).catch(toaster.catchGrpc)
    }

    function handleDrop() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Drop HAProxy instance?"
                message={`Drop ${name} and its configuration? This action cannot be undone.`}
                confirmLabel="Drop"
                onConfirm={handleConfirmDrop}
                onClose={CloseDialog}
            />,
        )
    }

    return (
        <div className={cn(cls.Row, isSelected && cls.Selected)}>
            <StatusDot status={dotStatus} tooltip={status}/>
            <Link className={cls.Name} to={Routes.Service + "/" + HAPROXY_NAME_PREFIX + name}>{name}</Link>
            <span className={cls.Cell}>{instance.environment || "—"}</span>
            <span className={cls.Cell}>{instance.port || "—"}</span>
            <span className={cls.Cell}>{instance.masterPgInstance || "—"}</span>
            <div className={cls.Actions}>
                <Button sm variant={isSelected ? "primary" : "secondary"} onClick={handleSelect}>Manage</Button>
                <Button sm variant="danger" onClick={handleDrop} disabled={dropInstance.isPending}>Drop</Button>
            </div>
        </div>
    )
}
