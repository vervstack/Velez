import {useMemo} from "react"

import cls from "@/pages/s3/components/S3KeysTab/S3KeysTab.module.css"
import type {S3Key} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useListS3KeysQuery} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import CreateS3KeyDialog from "@/dialogs/CreateS3KeyDialog/CreateS3KeyDialog.tsx"
import S3KeyRow from "@/pages/s3/components/S3KeyRow/S3KeyRow.tsx"
import S3ListSkeleton from "@/pages/s3/components/S3ListSkeleton/S3ListSkeleton.tsx"

const COLUMNS = ["Name", "Access key id", "Buckets", "Owner", "Secret", ""]

function renderColumnHeader(label: string, i: number) {
    return <span key={i} className={cls.HeaderCell}>{label}</span>
}

interface Props {
    instanceName: string
}

export default function S3KeysTab({instanceName}: Props) {
    const {OpenDialog} = useDialog()
    const keysQuery = useListS3KeysQuery(instanceName)

    const keys = useMemo(
        () => [...(keysQuery.data?.keys ?? [])].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? "")),
        [keysQuery.data]
    )

    function handleCreate() {
        OpenDialog(<CreateS3KeyDialog instanceName={instanceName}/>)
    }

    function handleRetry() {
        keysQuery.refetch()
    }

    function renderRow(key: S3Key) {
        return <S3KeyRow key={key.accessKeyId} instanceName={instanceName} s3Key={key}/>
    }

    function renderContent() {
        if (keysQuery.isLoading) return <S3ListSkeleton/>
        if (keysQuery.isError) {
            return <QueryErrorState message="Failed to load keys." onRetry={handleRetry}/>
        }
        if (keys.length === 0) {
            return <div className={cls.EmptyMessage}>No keys in this instance.</div>
        }
        return (
            <div className={cls.Table}>
                <div className={cls.TableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {keys.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.S3KeysTabContainer}>
            <div className={cls.Toolbar}>
                <Button variant="primary" sm onClick={handleCreate}>Create key</Button>
            </div>
            {renderContent()}
        </div>
    )
}
