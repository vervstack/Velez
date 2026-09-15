import {useState} from "react"

import cls from "@/pages/container-registry/components/RegistryInstanceCredentials/RegistryInstanceCredentials.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import {GetRegistryInstanceCredentialsQuery} from "@/processes/queries/registry_instances.ts"

interface Props {
    name: string
    username: string
}

const MASKED_PASSWORD = "••••••••••••"

export default function RegistryInstanceCredentials({name, username}: Props) {
    const [revealed, setRevealed] = useState(false)
    const toaster = useToaster()

    const credentialsQuery = GetRegistryInstanceCredentialsQuery(name)

    function copyToClipboard(text: string) {
        navigator.clipboard.writeText(text)
            .then(function onCopied() {
                toaster.bake({title: "Docker login command copied", description: name, level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    function buildDockerLoginCommand(registryUrl: string, password: string): string {
        return `docker login ${registryUrl} -u ${username} -p ${password}`
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

    function handleCopyDockerLogin() {
        if (credentialsQuery.data?.password && credentialsQuery.data?.registryUrl) {
            copyToClipboard(buildDockerLoginCommand(credentialsQuery.data.registryUrl, credentialsQuery.data.password))
            return
        }

        credentialsQuery.refetch()
            .then(function onFetched(result) {
                if (result.data?.password && result.data?.registryUrl) {
                    copyToClipboard(buildDockerLoginCommand(result.data.registryUrl, result.data.password))
                }
            })
            .catch(toaster.catchGrpc)
    }

    const passwordValue = revealed && credentialsQuery.data?.password
        ? credentialsQuery.data.password
        : MASKED_PASSWORD

    return (
        <div className={cls.RegistryInstanceCredentialsContainer}>
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
                <Button sm onClick={handleCopyDockerLogin}>
                    Copy docker login command
                </Button>
            </div>
        </div>
    )
}
