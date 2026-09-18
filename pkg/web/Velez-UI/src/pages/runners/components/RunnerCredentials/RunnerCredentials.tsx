import {useState} from "react"

import cls from "@/pages/runners/components/RunnerCredentials/RunnerCredentials.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import {GetRunnerCredentialsQuery} from "@/processes/queries/runners.ts"

interface Props {
    name: string
}

const MASKED_TOKEN = "••••••••••••"

export default function RunnerCredentials({name}: Props) {
    const [revealed, setRevealed] = useState(false)
    const toaster = useToaster()

    const credentialsQuery = GetRunnerCredentialsQuery(name)

    function copyToClipboard(text: string, description: string) {
        navigator.clipboard.writeText(text)
            .then(function onCopied() {
                toaster.bake({title: description, description: name, level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    function handleCredentialsUnavailable() {
        toaster.bake({
            title: "Token unavailable",
            description: `${name}'s credentials were lost, likely after a server restart. `
                + "Recreate the runner to get a new token.",
            level: "Error",
        })
    }

    function handleToggleReveal() {
        if (revealed) {
            setRevealed(false)
            return
        }
        setRevealed(true)
        if (!credentialsQuery.data) {
            credentialsQuery.refetch().catch(handleCredentialsUnavailable)
        }
    }

    function handleCopyToken() {
        if (credentialsQuery.data?.token) {
            copyToClipboard(credentialsQuery.data.token, "Token copied")
            return
        }

        credentialsQuery.refetch()
            .then(function onFetched(result) {
                if (result.data?.token) {
                    copyToClipboard(result.data.token, "Token copied")
                }
            })
            .catch(handleCredentialsUnavailable)
    }

    function handleCopyRegisterCommand() {
        if (credentialsQuery.data?.registerCommand) {
            copyToClipboard(credentialsQuery.data.registerCommand, "Register command copied")
        }
    }

    const tokenValue = revealed && credentialsQuery.data?.token
        ? credentialsQuery.data.token
        : MASKED_TOKEN

    const registerCommand = credentialsQuery.data?.registerCommand

    return (
        <div className={cls.RunnerCredentialsContainer}>
            <div className={cls.TokenRow}>
                <span className={cls.Label}>Token</span>
                <span className={cls.Value}>
                    {credentialsQuery.isFetching ? "loading…" : tokenValue}
                </span>
                <Button sm onClick={handleToggleReveal}>
                    {revealed ? "Hide" : "Reveal"}
                </Button>
            </div>
            {revealed && registerCommand && (
                <div className={cls.TokenRow}>
                    <span className={cls.Label}>Register</span>
                    <span className={cls.Value}>{registerCommand}</span>
                </div>
            )}
            <div className={cls.ActionsRow}>
                <Button sm onClick={handleCopyToken}>
                    Copy token
                </Button>
                {revealed && registerCommand && (
                    <Button sm onClick={handleCopyRegisterCommand}>
                        Copy register command
                    </Button>
                )}
            </div>
        </div>
    )
}
