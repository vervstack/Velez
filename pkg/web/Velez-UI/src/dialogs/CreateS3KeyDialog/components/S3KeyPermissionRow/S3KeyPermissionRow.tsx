import {Dropdown, DropdownOption} from "@vervstack/chures"

import cls from "@/dialogs/CreateS3KeyDialog/components/S3KeyPermissionRow/S3KeyPermissionRow.module.css"
import Button from "@/components/base/Button.tsx"
import Checkbox from "@/components/base/Checkbox.tsx"
import type {KeyPermissionRow} from "@/dialogs/CreateS3KeyDialog/processes/buildCreateS3KeyRequest.ts"

interface Props {
    row: KeyPermissionRow
    bucketOptions: DropdownOption[]
    isLoadingBuckets: boolean

    onChange(row: KeyPermissionRow): void

    onRemove(id: number): void
}

export default function S3KeyPermissionRow({row, bucketOptions, isLoadingBuckets, onChange, onRemove}: Props) {
    function handleBucketChange(ids: string[]) {
        onChange({...row, bucketName: ids[0] ?? ""})
    }

    function handleReadChange(isRead: boolean) {
        onChange({...row, isRead})
    }

    function handleWriteChange(isWrite: boolean) {
        onChange({...row, isWrite})
    }

    function handleOwnerChange(isOwner: boolean) {
        onChange({...row, isOwner})
    }

    function handleRemove() {
        onRemove(row.id)
    }

    return (
        <div className={cls.S3KeyPermissionRowContainer}>
            <Dropdown
                label="Bucket"
                placeholder="Pick a bucket"
                options={bucketOptions}
                value={row.bucketName ? [row.bucketName] : []}
                onChange={handleBucketChange}
                isLoading={isLoadingBuckets}
                portal
            />
            <div className={cls.FlagsWrapper}>
                <Checkbox label="Read" checked={row.isRead} onChange={handleReadChange}/>
                <Checkbox label="Write" checked={row.isWrite} onChange={handleWriteChange}/>
                <Checkbox label="Owner" checked={row.isOwner} onChange={handleOwnerChange}/>
                <Button sm variant="ghost" onClick={handleRemove}>✕</Button>
            </div>
        </div>
    )
}
