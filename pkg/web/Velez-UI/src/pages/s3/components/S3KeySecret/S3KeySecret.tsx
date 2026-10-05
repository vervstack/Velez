import {useState} from "react"

import cls from "@/pages/s3/components/S3KeySecret/S3KeySecret.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {GetS3KeyCredentialsQuery} from "@/processes/queries/s3.ts"
import Button from "@/components/base/Button.tsx"

const MASKED_SECRET = "••••••••••••"

interface Props {
    instanceName: string
    accessKeyId: string
}

export default function S3KeySecret({instanceName, accessKeyId}: Props) {
    const [isRevealed, setIsRevealed] = useState(false)
    const toaster = useToaster()

    const credentialsQuery = GetS3KeyCredentialsQuery(instanceName, accessKeyId)
    const secret = credentialsQuery.data?.secretAccessKey

    function copyToClipboard(value: string) {
        navigator.clipboard.writeText(value)
            .then(function onCopied() {
                toaster.bake({title: "Secret copied", description: accessKeyId, level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    function handleToggleReveal() {
        if (isRevealed) {
            setIsRevealed(false)
            return
        }
        setIsRevealed(true)
        if (!secret) {
            credentialsQuery.refetch().catch(toaster.catchGrpc)
        }
    }

    function handleCopy() {
        if (secret) {
            copyToClipboard(secret)
            return
        }

        credentialsQuery.refetch()
            .then(function onFetched(result) {
                if (result.data?.secretAccessKey) copyToClipboard(result.data.secretAccessKey)
            })
            .catch(toaster.catchGrpc)
    }

    const shownSecret = isRevealed && secret ? secret : MASKED_SECRET

    return (
        <div className={cls.S3KeySecretContainer}>
            <span className={cls.Value}>{credentialsQuery.isFetching ? "loading…" : shownSecret}</span>
            <div className={cls.Actions}>
                <Button sm onClick={handleToggleReveal}>{isRevealed ? "Hide" : "Reveal"}</Button>
                <Button sm onClick={handleCopy}>Copy</Button>
            </div>
        </div>
    )
}
