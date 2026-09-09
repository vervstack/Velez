import {useState} from "react"

import cls from "@/pages/postgres/components/PgInstanceCredentials/PgInstanceCredentials.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
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

    function copyToClipboard(dsn: string) {
        navigator.clipboard.writeText(dsn)
            .then(function onCopied() {
                toaster.bake({title: "DSN copied", description: name, level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    function handleToggleReveal() {
        if (revealed) {
            setRevealed(false)
            return
        }
        setRevealed(true)
        if (!credentialsQuery.data) {
            credentialsQuery.refetch().catch(toaster.catchGrpc)
        }
    }

    function handleCopyDsn() {
        if (credentialsQuery.data?.dsn) {
            copyToClipboard(credentialsQuery.data.dsn)
            return
        }

        credentialsQuery.refetch()
            .then(function onFetched(result) {
                if (result.data?.dsn) copyToClipboard(result.data.dsn)
            })
            .catch(toaster.catchGrpc)
    }

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
                    {credentialsQuery.isFetching ? "loading…" : passwordValue}
                </span>
                <Button sm onClick={handleToggleReveal}>
                    {revealed ? "Hide" : "Reveal"}
                </Button>
            </div>
            <div className={cls.ActionsRow}>
                <Button sm onClick={handleCopyDsn}>
                    Copy DSN
                </Button>
            </div>
        </div>
    )
}
