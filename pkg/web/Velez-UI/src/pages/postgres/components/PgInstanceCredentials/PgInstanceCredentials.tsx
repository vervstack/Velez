import {useEffect, useState} from "react"

import cls from "@/pages/postgres/components/PgInstanceCredentials/PgInstanceCredentials.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import PgConnectionRow from "@/pages/postgres/components/PgConnectionRow/PgConnectionRow.tsx"
import {GetPgInstanceCredentialsQuery} from "@/processes/queries/pg_instances.ts"

interface Props {
    name: string
    dbName: string
    username: string
}

const MASKED_PASSWORD = "••••••••••••"

export default function PgInstanceCredentials({name, dbName, username}: Props) {
    const [revealed, setRevealed] = useState(false)
    const toaster = useToaster()

    const credentialsQuery = GetPgInstanceCredentialsQuery(name)
    useEffect(() => {
        if (credentialsQuery.error) toaster.catchGrpc(credentialsQuery.error)
    }, [credentialsQuery.error])

    function copyToClipboard(dsn?: string) {
        if (!dsn) return

        navigator.clipboard.writeText(dsn)
            .then(function onCopied() {
                toaster.bake({title: "DSN copied", description: name, level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    function handleToggleReveal() {
        setRevealed(!revealed)
    }

    function handleCopyInNetwork() {
        copyToClipboard(credentialsQuery.data?.inNetwork?.dsn)
    }

    function handleCopyExternal() {
        copyToClipboard(credentialsQuery.data?.external?.dsn)
    }

    const inNetwork = credentialsQuery.data?.inNetwork
    const external = credentialsQuery.data?.external

    const passwordValue = revealed && credentialsQuery.data?.password
        ? credentialsQuery.data.password
        : MASKED_PASSWORD

    return (
        <div className={cls.PgInstanceCredentialsContainer}>
            <div className={cls.TokenRow}>
                <span className={cls.Label}>Database</span>
                <span className={cls.Value}>{dbName}</span>
            </div>
            <div className={cls.TokenRow}>
                <span className={cls.Label}>Username</span>
                <span className={cls.Value}>{username}</span>
            </div>
            <div className={cls.TokenRow}>
                <span className={cls.Label}>Password</span>
                <span className={cls.Value}>
                    {credentialsQuery.isLoading ? "loading…" : passwordValue}
                </span>
                <Button sm onClick={handleToggleReveal}>
                    {revealed ? "Hide" : "Reveal"}
                </Button>
            </div>
            <PgConnectionRow
                label="In-network"
                maskedDsn={inNetwork?.maskedDsn}
                isLoading={credentialsQuery.isLoading}
                onCopy={handleCopyInNetwork}
            />
            {external && (
                <PgConnectionRow
                    label="External"
                    maskedDsn={external.maskedDsn}
                    isLoading={credentialsQuery.isLoading}
                    onCopy={handleCopyExternal}
                />
            )}
        </div>
    )
}
