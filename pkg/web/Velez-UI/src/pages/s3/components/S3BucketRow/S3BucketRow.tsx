import {ConfirmDialog} from "@vervstack/chures"

import cls from "@/pages/s3/components/S3BucketRow/S3BucketRow.module.css"
import type {S3Bucket} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {formatBytes} from "@/processes/mappings/s3.ts"
import {DeleteS3BucketMutation} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import S3BucketAccessList from "@/pages/s3/components/S3BucketAccessList/S3BucketAccessList.tsx"

interface Props {
    instanceName: string
    bucket: S3Bucket
}

export default function S3BucketRow({instanceName, bucket}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const deleteBucket = DeleteS3BucketMutation()

    const name = bucket.name ?? ""

    function handleConfirmDelete() {
        deleteBucket.mutateAsync({instanceName, bucketName: name})
            .then(function handleDeleted() {
                toaster.bake({title: "Bucket deleted", description: name, level: "Info"})
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
    }

    function handleDelete() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Delete bucket?"
                message={`${name} and all of its objects will be removed. This cannot be undone.`}
                confirmLabel="Delete"
                onConfirm={handleConfirmDelete}
                onClose={CloseDialog}
            />
        )
    }

    return (
        <div className={cls.S3BucketRowContainer}>
            <div className={cls.Row}>
                <span className={cls.Name}>{name}</span>
                <span className={cls.Cell}>{bucket.objectCount ?? "0"}</span>
                <span className={cls.Cell}>{formatBytes(bucket.bytes)}</span>
                <span className={cls.Cell}>{bucket.ownerService || "-"}</span>
                <S3BucketAccessList instanceName={instanceName} access={bucket.access ?? []}/>
                <div className={cls.Actions}>
                    <Button sm variant="danger" onClick={handleDelete} disabled={deleteBucket.isPending}>
                        Delete
                    </Button>
                </div>
            </div>
        </div>
    )
}
