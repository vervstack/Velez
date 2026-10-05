import {ConfirmDialog} from "@vervstack/chures"

import cls from "@/pages/s3/components/S3KeyRow/S3KeyRow.module.css"
import type {S3Key} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {formatKeyBuckets} from "@/processes/mappings/s3.ts"
import {DeleteS3KeyMutation} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import S3KeySecret from "@/pages/s3/components/S3KeySecret/S3KeySecret.tsx"

interface Props {
    instanceName: string
    s3Key: S3Key
}

export default function S3KeyRow({instanceName, s3Key}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const deleteKey = DeleteS3KeyMutation()

    const name = s3Key.name ?? ""
    const accessKeyId = s3Key.accessKeyId ?? ""

    function handleConfirmDelete() {
        deleteKey.mutateAsync({instanceName, accessKeyId})
            .then(function handleDeleted() {
                toaster.bake({title: "Key deleted", description: name || accessKeyId, level: "Info"})
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
    }

    function handleDelete() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Delete key?"
                message={`${name || accessKeyId} will be revoked. Services using it lose access.`}
                confirmLabel="Delete"
                onConfirm={handleConfirmDelete}
                onClose={CloseDialog}
            />
        )
    }

    return (
        <div className={cls.S3KeyRowContainer}>
            <div className={cls.Row}>
                <span className={cls.Name}>{name || "-"}</span>
                <span className={cls.Cell}>{accessKeyId}</span>
                <span className={cls.Cell}>{formatKeyBuckets(s3Key.access ?? [])}</span>
                <span className={cls.Cell}>{s3Key.ownerService || "-"}</span>
                <S3KeySecret instanceName={instanceName} accessKeyId={accessKeyId}/>
                <div className={cls.Actions}>
                    <Button sm variant="danger" onClick={handleDelete} disabled={deleteKey.isPending}>
                        Delete
                    </Button>
                </div>
            </div>
        </div>
    )
}
