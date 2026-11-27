import {Button, Dropdown, Input} from "@vervstack/chures"
import {type FormEvent, useId, useMemo, useRef, useState} from "react"

import cls from "@/dialogs/TailscaleSidecarDialog/TailscaleSidecarDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useConnectServiceMutation, useListVcnNetworksQuery} from "@/processes/queries/vcn.ts"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"

const DEFAULT_NETWORK_OPTION = {id: "legacy", name: "Default (legacy)"}

interface Props {
    serviceName: string
}

export default function TailscaleSidecarDialog({serviceName}: Props) {
    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const {connectService, isPending} = useConnectServiceMutation(serviceName)
    const networksQuery = useListVcnNetworksQuery()
    const form = useRef<HTMLFormElement>(null)
    const authKeyHintId = useId()
    const [authKey, setAuthKey] = useState("")
    const [loginServerUrl, setLoginServerUrl] = useState("")
    const [hostname, setHostname] = useState("")
    const [networkId, setNetworkId] = useState("")
    const options = useMemo(() => [
        DEFAULT_NETWORK_OPTION,
        ...(networksQuery.data ?? []).flatMap(({id, name}) => id ? [{id, name: name || id}] : []),
    ], [networksQuery.data])
    const selectedNetwork = networksQuery.data?.find((network) => network.id === networkId)
    const hasNamedNetwork = Boolean(networkId)

    function handleNetworkChange(ids: string[]) {
        if (isPending) return
        const id = ids[0]
        setNetworkId(id && id !== DEFAULT_NETWORK_OPTION.id ? id : "")
    }

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()

        connectService({
            serviceName,
            networkId,
            authKey: hasNamedNetwork ? undefined : authKey.trim() || undefined,
            loginServerUrl: hasNamedNetwork ? undefined : loginServerUrl.trim() || undefined,
            hostname: hostname.trim() || undefined,
        })?.then(function handleConnected() {
            toaster.bake({title: "Tailscale sidecar started", description: serviceName, level: "Info"})
            if (form.current) CloseDialog()
        }).catch(toaster.catchGrpc)
    }

    return (
        <DialogShell title="Connect to Tailscale" onClose={CloseDialog} isFlush>
            <form ref={form} className={cls.TailscaleSidecarDialogContainer} onSubmit={handleSubmit}>
                <Dropdown
                    className={cls.Field}
                    label="VCN network"
                    options={options}
                    value={[networkId || DEFAULT_NETWORK_OPTION.id]}
                    onChange={handleNetworkChange}
                    isLoading={networksQuery.isLoading}
                    portal
                />
                {networksQuery.isError && (
                    <div className={cls.Field}>
                        <QueryErrorState message="Failed to load VCN networks." onRetry={networksQuery.refetch}/>
                    </div>
                )}
                <Input
                    className={cls.Field}
                    label="Auth key (optional)"
                    aria-label="Auth key (optional)"
                    type="password"
                    autoComplete="off"
                    aria-describedby={authKeyHintId}
                    value={authKey}
                    setValue={setAuthKey}
                    disabled={isPending || hasNamedNetwork}
                />
                <small className={cls.Hint} id={authKeyHintId}>
                    {hasNamedNetwork
                        ? "The selected network provides the auth key and login server URL."
                        : "Leave empty to use cluster Headscale"}
                </small>
                <Input
                    className={cls.Field}
                    label="Login server URL (optional)"
                    aria-label="Login server URL (optional)"
                    value={hasNamedNetwork ? selectedNetwork?.url ?? "" : loginServerUrl}
                    setValue={setLoginServerUrl}
                    disabled={isPending || hasNamedNetwork}
                />
                <Input
                    className={cls.Field}
                    label="Hostname (optional)"
                    aria-label="Hostname (optional)"
                    value={hostname}
                    setValue={setHostname}
                    disabled={isPending}
                />
                <Button className={cls.Cancel} variant="secondary" type="button" onClick={CloseDialog}>
                    Cancel
                </Button>
                <Button variant="primary" type="submit" disabled={isPending}>
                    {isPending ? "Connecting…" : "Connect"}
                </Button>
            </form>
        </DialogShell>
    )
}
