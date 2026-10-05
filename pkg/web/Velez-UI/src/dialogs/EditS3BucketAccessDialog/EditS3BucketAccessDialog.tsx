import {useState} from "react"

import cls from "@/dialogs/EditS3BucketAccessDialog/EditS3BucketAccessDialog.module.css"
import type {S3BucketAccess, SetS3BucketAccessRequest} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {SetS3BucketAccessMutation} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import Checkbox from "@/components/base/Checkbox.tsx"

interface Props {
    instanceName: string
    access: S3BucketAccess
}

export default function EditS3BucketAccessDialog({instanceName, access}: Props) {
    const [isRead, setIsRead] = useState(access.isRead ?? false)
    const [isWrite, setIsWrite] = useState(access.isWrite ?? false)
    const [isOwner, setIsOwner] = useState(access.isOwner ?? false)

    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const setAccess = SetS3BucketAccessMutation()

    const keyLabel = access.keyName || access.accessKeyId

    function handleSave() {
        const req: SetS3BucketAccessRequest = {
            instanceName,
            access: {...access, isRead, isWrite, isOwner},
        }

        setAccess.mutateAsync(req)
            .then(function handleSaved() {
                const description = `${keyLabel} on ${access.bucketName}`
                toaster.bake({title: "Access updated", description, level: "Info"})
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.EditS3BucketAccessDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Access of {keyLabel} to {access.bucketName}</h2>
            </div>
            <div className={cls.Content}>
                <div className={cls.FieldsWrapper}>
                    <Checkbox label="Read" checked={isRead} onChange={setIsRead}/>
                    <Checkbox label="Write" checked={isWrite} onChange={setIsWrite}/>
                    <Checkbox label="Owner" checked={isOwner} onChange={setIsOwner}/>
                </div>
                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={setAccess.isPending}>
                        Cancel
                    </Button>
                    <Button variant="primary" onClick={handleSave} disabled={setAccess.isPending}>
                        {setAccess.isPending ? "Saving…" : "Save"}
                    </Button>
                </div>
            </div>
        </div>
    )
}
