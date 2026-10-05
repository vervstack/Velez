import {useState} from "react"

import cls from "@/pages/s3/components/S3CredentialRow/S3CredentialRow.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"

const MASKED_VALUE = "••••••••••••"

interface Props {
    label: string
    value?: string
    isSecret?: boolean
}

export default function S3CredentialRow({label, value, isSecret = false}: Props) {
    const [isRevealed, setIsRevealed] = useState(false)
    const toaster = useToaster()

    function handleToggleReveal() {
        setIsRevealed(!isRevealed)
    }

    function handleCopy() {
        if (!value) return

        navigator.clipboard.writeText(value)
            .then(function onCopied() {
                toaster.bake({title: `${label} copied`, description: "", level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    if (!value) return null

    return (
        <div className={cls.S3CredentialRowContainer}>
            <span className={cls.Label}>{label}</span>
            <span className={cls.Value}>{isSecret && !isRevealed ? MASKED_VALUE : value}</span>
            <div className={cls.Actions}>
                {isSecret && (
                    <Button sm onClick={handleToggleReveal}>{isRevealed ? "Hide" : "Reveal"}</Button>
                )}
                <Button sm onClick={handleCopy}>Copy</Button>
            </div>
        </div>
    )
}
