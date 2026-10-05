import {Link} from "react-router-dom"
import {ConfirmDialog} from "@vervstack/chures"
import cn from "classnames"

import cls from "@/pages/s3/components/S3InstanceRow/S3InstanceRow.module.css"
import type {S3Instance} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {Routes} from "@/app/router/Routes.ts"
import {mapS3InstanceStatus, s3ServiceName, s3WebUiLink} from "@/processes/mappings/s3.ts"
import {DropS3InstanceMutation} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import StatusDot from "@/components/base/StatusDot.tsx"

interface Props {
    instance: S3Instance
    isSelected: boolean

    onSelect(name: string): void
}

export default function S3InstanceRow({instance, isSelected, onSelect}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const dropInstance = DropS3InstanceMutation()

    const name = instance.name ?? ""
    const webUiLink = s3WebUiLink(instance)

    function handleSelect() {
        onSelect(name)
    }

    function handleConfirmDrop() {
        dropInstance.mutateAsync(name)
            .then(function handleDropped() {
                toaster.bake({title: "S3 instance dropped", description: name, level: "Info"})
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
    }

    function handleDrop() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Drop S3 instance?"
                message={`${name} and every bucket and key in it will be removed. This cannot be undone.`}
                confirmLabel="Drop"
                onConfirm={handleConfirmDrop}
                onClose={CloseDialog}
            />
        )
    }

    return (
        <div className={cn(cls.S3InstanceRowContainer, isSelected && cls.Selected)}>
            <div className={cls.Row}>
                <StatusDot status={mapS3InstanceStatus(instance.status)} pulse/>
                <Link className={cls.Name} to={Routes.Service + "/" + s3ServiceName(name)}>{name}</Link>
                <span className={cls.Cell}>{instance.environment || "-"}</span>
                <span className={cls.Cell}>{instance.s3Port || "-"}</span>
                <span className={cls.Cell}>
                    {webUiLink
                        ? <a className={cls.Link} href={webUiLink} target="_blank" rel="noreferrer">Open ↗</a>
                        : "-"}
                </span>
                <span className={cls.Cell}>{instance.region || "-"}</span>
                <div className={cls.Actions}>
                    <Button sm variant={isSelected ? "primary" : "secondary"} onClick={handleSelect}>
                        Manage
                    </Button>
                    <Button sm variant="danger" onClick={handleDrop} disabled={dropInstance.isPending}>
                        Drop
                    </Button>
                </div>
            </div>
        </div>
    )
}
