import cls from "@/pages/s3/components/S3BucketAccessList/S3BucketAccessList.module.css"
import type {S3BucketAccess} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {formatAccessFlags} from "@/processes/mappings/s3.ts"
import Button from "@/components/base/Button.tsx"
import EditS3BucketAccessDialog from "@/dialogs/EditS3BucketAccessDialog/EditS3BucketAccessDialog.tsx"

interface Props {
    instanceName: string
    access: S3BucketAccess[]
}

export default function S3BucketAccessList({instanceName, access}: Props) {
    const {OpenDialog} = useDialog()

    function renderEntry(entry: S3BucketAccess) {
        function handleEdit() {
            OpenDialog(<EditS3BucketAccessDialog instanceName={instanceName} access={entry}/>)
        }

        return (
            <div key={entry.accessKeyId} className={cls.Entry}>
                <span className={cls.EntryText}>
                    {entry.keyName || entry.accessKeyId}: {formatAccessFlags(entry)}
                </span>
                <Button sm onClick={handleEdit}>Edit</Button>
            </div>
        )
    }

    return (
        <div className={cls.S3BucketAccessListContainer}>
            {access.length === 0 && <span className={cls.EmptyText}>-</span>}
            {access.map(renderEntry)}
        </div>
    )
}
