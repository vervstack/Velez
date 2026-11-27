import {useState} from "react"
import {Button} from "@vervstack/chures"

import type {NetworkApiKey} from "@/app/api/velez/verv_closed_network.pb"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import VcnApiKeyRow from "@/pages/vcn/components/VcnApiKeyRow/VcnApiKeyRow.tsx"
import {useIssueNetworkApiKeyMutation} from "@/processes/queries/vcn.ts"
import cls from "@/pages/vcn/components/VcnApiKeys/VcnApiKeys.module.css"

interface Props {
    networkId: string
    apiKeys?: NetworkApiKey[]
}

export default function VcnApiKeys({networkId, apiKeys = []}: Props) {
    const [secret, setSecret] = useState("")
    const toaster = useToaster()
    const issueKey = useIssueNetworkApiKeyMutation(networkId)

    function handleIssue() {
        issueKey.mutateAsync()
            .then(function handleIssued(response) {
                if (!response.key) throw new Error("Headscale did not return an API key.")
                setSecret(response.key)
            })
            .catch(toaster.catchGrpc)
            .finally(issueKey.reset)
    }

    function handleCopy() {
        Promise.resolve()
            .then(function copySecret() {
                if (!navigator.clipboard) throw new Error("Clipboard is unavailable. Select the key to copy it.")
                return navigator.clipboard.writeText(secret)
            })
            .then(function handleCopied() {
                toaster.bake({title: "API key copied", description: "", level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    function handleClose() {
        setSecret("")
    }

    function renderSecret() {
        if (!secret) return null

        return (
            <div className={cls.Secret}>
                <p className={cls.SecretNote} role="status">Copy this API key now. It will only be shown once.</p>
                <code className={cls.SecretValue}>{secret}</code>
                <div className={cls.ActionsWrapper}>
                    <Button onClick={handleCopy}>Copy API key</Button>
                    <Button variant="ghost" onClick={handleClose}>Close</Button>
                </div>
            </div>
        )
    }

    function renderKeyRow(apiKey: NetworkApiKey) {
        return <VcnApiKeyRow key={apiKey.prefix} networkId={networkId} apiKey={apiKey}/>
    }

    function renderTableHeading() {
        return (
            <thead>
                <tr>
                    <th scope="col">Prefix</th>
                    <th scope="col">Created</th>
                    <th scope="col">Expires</th>
                    <th scope="col">Last seen</th>
                    <th scope="col">Actions</th>
                </tr>
            </thead>
        )
    }

    function renderApiKeys() {
        if (apiKeys.length === 0) return <p className={cls.EmptyMessage}>No API keys.</p>

        return (
            <table className={cls.Table} aria-label="Network API keys">
                {renderTableHeading()}
                <tbody>{apiKeys.map(renderKeyRow)}</tbody>
            </table>
        )
    }

    return (
        <section className={cls.VcnApiKeysContainer}>
            <header className={cls.Toolbar}>
                <h2 className={cls.Title}>API keys</h2>
                <Button variant="primary" onClick={handleIssue} disabled={issueKey.isPending || !!secret}>
                    {issueKey.isPending ? "Issuing…" : "Issue API key"}
                </Button>
            </header>
            {renderSecret()}
            <div className={cls.TableWrapper}>{renderApiKeys()}</div>
        </section>
    )
}
