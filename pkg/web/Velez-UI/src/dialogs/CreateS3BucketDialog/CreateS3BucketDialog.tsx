import {useState} from "react"
import {Dropdown, DropdownOption, parseGrpcError} from "@vervstack/chures"

import cls from "@/dialogs/CreateS3BucketDialog/CreateS3BucketDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateS3BucketMutation} from "@/processes/queries/s3.ts"
import {useListServicesQuery} from "@/processes/queries/services.ts"
import Button from "@/components/base/Button.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import Input from "@/components/base/Input.tsx"
import {buildCreateS3BucketRequest} from "@/dialogs/CreateS3BucketDialog/processes/buildCreateS3BucketRequest.ts"

interface Props {
    instanceName: string
}

export default function CreateS3BucketDialog({instanceName}: Props) {
    const [bucketName, setBucketName] = useState("")
    const [ownerService, setOwnerService] = useState("")

    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const servicesQuery = useListServicesQuery()
    const createBucket = CreateS3BucketMutation()

    const serviceOptions: DropdownOption[] = (servicesQuery.data?.services ?? []).map((s) => ({
        id: s.name ?? "",
        name: s.name ?? "",
    }))

    const req = buildCreateS3BucketRequest(instanceName, bucketName, ownerService)

    function handleError(err: unknown) {
        toaster.catchGrpc(parseGrpcError(err))
    }

    function handleOwnerChange(ids: string[]) {
        setOwnerService(ids[0] ?? "")
    }

    function handleCreate() {
        if (!req) return

        createBucket.mutateAsync(req)
            .then(function handleCreated() {
                toaster.bake({title: "Bucket created", description: req.bucketName ?? "", level: "Info"})
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.CreateS3BucketDialogContainer}>
            <DialogShell title={`Create bucket in ${instanceName}`} onClose={CloseDialog}>
                <div className={cls.FieldsWrapper}>
                    <Input label="Bucket name" inputValue={bucketName} onChange={setBucketName}/>
                    <Dropdown
                        label="Owner service (optional)"
                        placeholder="None"
                        options={serviceOptions}
                        value={ownerService ? [ownerService] : []}
                        onChange={handleOwnerChange}
                        isLoading={servicesQuery.isLoading}
                        onError={handleError}
                        portal
                    />
                    <p className={cls.Description}>
                        An owner service gets a read/write key named after the service and the bucket.
                    </p>
                </div>
                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={createBucket.isPending}>
                        Cancel
                    </Button>
                    <Button variant="primary" onClick={handleCreate} disabled={createBucket.isPending || !req}>
                        {createBucket.isPending ? "Creating…" : "Create"}
                    </Button>
                </div>
            </DialogShell>
        </div>
    )
}
