import {Dropdown, DropdownOption} from "@vervstack/chures"

import cls
    from "@/dialogs/CreateServiceDialog/screens/RegistryScreen/components/RegistryStorageFields/RegistryStorageFields.module.css"
import {useListS3InstancesQuery} from "@/processes/queries/s3.ts"
import Choice from "@/components/base/Choice.tsx"
import Input from "@/components/base/Input.tsx"
import type {
    RegistryStorageKind,
    RegistryStorageState,
} from "@/dialogs/CreateServiceDialog/screens/RegistryScreen/processes/registryStorage.ts"

interface Props {
    value: RegistryStorageState
    bucketPlaceholder: string

    onChange(next: RegistryStorageState): void
}

export default function RegistryStorageFields({value, bucketPlaceholder, onChange}: Props) {
    const instancesQuery = useListS3InstancesQuery()

    const instanceOptions: DropdownOption[] = (instancesQuery.data?.instances ?? []).map((instance) => ({
        id: instance.name ?? "",
        name: instance.name ?? "",
    }))
    const hasNoInstances = !instancesQuery.isLoading && instanceOptions.length === 0

    function handleKindChange(kind: RegistryStorageKind) {
        onChange({...value, kind})
    }

    function handleLocalClick() {
        handleKindChange("local")
    }

    function handleS3Click() {
        handleKindChange("s3")
    }

    function handleInstanceChange(ids: string[]) {
        onChange({...value, s3Instance: ids[0] ?? ""})
    }

    function handleBucketChange(s3Bucket: string) {
        onChange({...value, s3Bucket})
    }

    return (
        <div className={cls.RegistryStorageFieldsContainer}>
            <span className={cls.FieldLabel}>Storage</span>
            <div className={cls.ChoiceRow}>
                <Choice title="Local volume" active={value.kind === "local"} onClick={handleLocalClick}/>
                <Choice title="S3 instance" active={value.kind === "s3"} onClick={handleS3Click}/>
            </div>

            {value.kind === "s3" && (
                <div className={cls.S3FieldsWrapper}>
                    <Dropdown
                        label="S3 instance"
                        placeholder="Pick an instance"
                        options={instanceOptions}
                        value={value.s3Instance ? [value.s3Instance] : []}
                        onChange={handleInstanceChange}
                        isLoading={instancesQuery.isLoading}
                        portal
                    />
                    {hasNoInstances && (
                        <span className={cls.Hint}>No S3 instances yet — create one on the S3 storage page.</span>
                    )}
                    <Input
                        label="Bucket (optional)"
                        inputValue={value.s3Bucket}
                        onChange={handleBucketChange}
                        placeholder={bucketPlaceholder}
                    />
                </div>
            )}
        </div>
    )
}
