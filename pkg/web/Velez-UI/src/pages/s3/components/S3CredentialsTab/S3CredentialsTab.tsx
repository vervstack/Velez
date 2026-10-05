import cls from "@/pages/s3/components/S3CredentialsTab/S3CredentialsTab.module.css"
import {GetS3InstanceCredentialsQuery} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import S3CredentialRow from "@/pages/s3/components/S3CredentialRow/S3CredentialRow.tsx"
import S3ListSkeleton from "@/pages/s3/components/S3ListSkeleton/S3ListSkeleton.tsx"

interface Props {
    instanceName: string
}

export default function S3CredentialsTab({instanceName}: Props) {
    const credentialsQuery = GetS3InstanceCredentialsQuery(instanceName)
    const credentials = credentialsQuery.data

    function handleLoad() {
        credentialsQuery.refetch()
    }

    if (credentialsQuery.isFetching && !credentials) return <S3ListSkeleton/>
    if (credentialsQuery.isError) {
        return <QueryErrorState message="Failed to load credentials." onRetry={handleLoad}/>
    }
    if (!credentials) {
        return (
            <div className={cls.S3CredentialsTabContainer}>
                <div className={cls.LoadWrapper}>
                    <span className={cls.Hint}>Credentials are only fetched on request.</span>
                    <Button variant="primary" sm onClick={handleLoad}>Load credentials</Button>
                </div>
            </div>
        )
    }

    return (
        <div className={cls.S3CredentialsTabContainer}>
            <div className={cls.Rows}>
                <S3CredentialRow label="Admin token" value={credentials.adminToken} isSecret/>
                <S3CredentialRow label="S3 endpoint" value={credentials.s3Endpoint}/>
                <S3CredentialRow label="Internal S3 endpoint" value={credentials.internalS3Endpoint}/>
                <S3CredentialRow label="Region" value={credentials.region}/>
                <S3CredentialRow label="Web UI url" value={credentials.webUiUrl}/>
                <S3CredentialRow label="Web UI user" value={credentials.webUiUsername}/>
                <S3CredentialRow label="Web UI password" value={credentials.webUiPassword} isSecret/>
            </div>
        </div>
    )
}
