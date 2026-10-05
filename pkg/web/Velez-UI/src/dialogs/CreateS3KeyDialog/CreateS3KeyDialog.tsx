import {useMemo, useState} from "react"
import {DropdownOption} from "@vervstack/chures"

import cls from "@/dialogs/CreateS3KeyDialog/CreateS3KeyDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateS3KeyMutation, useListS3BucketsQuery} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import Input from "@/components/base/Input.tsx"
import S3KeyPermissionRow from "@/dialogs/CreateS3KeyDialog/components/S3KeyPermissionRow/S3KeyPermissionRow.tsx"
import {
    buildCreateS3KeyRequest,
    KeyPermissionRow,
    newKeyPermissionRow,
} from "@/dialogs/CreateS3KeyDialog/processes/buildCreateS3KeyRequest.ts"

interface Props {
    instanceName: string
}

export default function CreateS3KeyDialog({instanceName}: Props) {
    const [keyName, setKeyName] = useState("")
    const [rows, setRows] = useState<KeyPermissionRow[]>([])
    const [nextRowId, setNextRowId] = useState(1)

    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const bucketsQuery = useListS3BucketsQuery(instanceName)
    const createKey = CreateS3KeyMutation()

    const bucketOptions = useMemo<DropdownOption[]>(
        () => (bucketsQuery.data?.buckets ?? []).map((bucket) => ({id: bucket.name ?? "", name: bucket.name ?? ""})),
        [bucketsQuery.data]
    )

    const req = buildCreateS3KeyRequest(instanceName, keyName, rows)

    function handleAddRow() {
        setRows([...rows, newKeyPermissionRow(nextRowId)])
        setNextRowId(nextRowId + 1)
    }

    function handleChangeRow(changed: KeyPermissionRow) {
        setRows(rows.map((row) => (row.id === changed.id ? changed : row)))
    }

    function handleRemoveRow(id: number) {
        setRows(rows.filter((row) => row.id !== id))
    }

    function handleCreate() {
        if (!req) return

        createKey.mutateAsync(req)
            .then(function handleCreated() {
                toaster.bake({title: "Key created", description: req.keyName ?? "", level: "Info"})
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
    }

    function renderRow(row: KeyPermissionRow) {
        return (
            <S3KeyPermissionRow
                key={row.id}
                row={row}
                bucketOptions={bucketOptions}
                isLoadingBuckets={bucketsQuery.isLoading}
                onChange={handleChangeRow}
                onRemove={handleRemoveRow}
            />
        )
    }

    return (
        <div className={cls.CreateS3KeyDialogContainer}>
            <DialogShell title={`Create key in ${instanceName}`} onClose={CloseDialog}>
                <div className={cls.FieldsWrapper}>
                    <Input label="Key name" inputValue={keyName} onChange={setKeyName}/>
                    {rows.map(renderRow)}
                    <Button sm onClick={handleAddRow}>Add bucket permission</Button>
                </div>
                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={createKey.isPending}>
                        Cancel
                    </Button>
                    <Button variant="primary" onClick={handleCreate} disabled={createKey.isPending || !req}>
                        {createKey.isPending ? "Creating…" : "Create"}
                    </Button>
                </div>
            </DialogShell>
        </div>
    )
}
