import {useMemo} from "react"

import cls from "@/pages/s3/components/S3BucketsTab/S3BucketsTab.module.css"
import type {S3Bucket} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useListS3BucketsQuery} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import CreateS3BucketDialog from "@/dialogs/CreateS3BucketDialog/CreateS3BucketDialog.tsx"
import S3BucketRow from "@/pages/s3/components/S3BucketRow/S3BucketRow.tsx"
import S3ListSkeleton from "@/pages/s3/components/S3ListSkeleton/S3ListSkeleton.tsx"

const COLUMNS = ["Name", "Objects", "Size", "Owner service", "Key access", ""]

function renderColumnHeader(label: string, i: number) {
    return <span key={i} className={cls.HeaderCell}>{label}</span>
}

interface Props {
    instanceName: string
}

export default function S3BucketsTab({instanceName}: Props) {
    const {OpenDialog} = useDialog()
    const bucketsQuery = useListS3BucketsQuery(instanceName)

    const buckets = useMemo(
        () => [...(bucketsQuery.data?.buckets ?? [])].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? "")),
        [bucketsQuery.data]
    )

    function handleCreate() {
        OpenDialog(<CreateS3BucketDialog instanceName={instanceName}/>)
    }

    function handleRetry() {
        bucketsQuery.refetch()
    }

    function renderRow(bucket: S3Bucket) {
        return <S3BucketRow key={bucket.id ?? bucket.name} instanceName={instanceName} bucket={bucket}/>
    }

    function renderContent() {
        if (bucketsQuery.isLoading) return <S3ListSkeleton/>
        if (bucketsQuery.isError) {
            return <QueryErrorState message="Failed to load buckets." onRetry={handleRetry}/>
        }
        if (buckets.length === 0) {
            return <div className={cls.EmptyMessage}>No buckets in this instance.</div>
        }
        return (
            <div className={cls.Table}>
                <div className={cls.TableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {buckets.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.S3BucketsTabContainer}>
            <div className={cls.Toolbar}>
                <Button variant="primary" sm onClick={handleCreate}>Create bucket</Button>
            </div>
            {renderContent()}
        </div>
    )
}
