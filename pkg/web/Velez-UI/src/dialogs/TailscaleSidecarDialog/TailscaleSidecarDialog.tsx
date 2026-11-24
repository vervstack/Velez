import {Button, Input} from "@vervstack/chures"
import {type FormEvent, useId, useRef, useState} from "react"

import cls from "@/dialogs/TailscaleSidecarDialog/TailscaleSidecarDialog.module.css"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useConnectServiceMutation} from "@/processes/queries/vcn.ts"

import DialogShell from "@/components/DialogShell/DialogShell.tsx"

interface Props {
    serviceName: string
}

export default function TailscaleSidecarDialog({serviceName}: Props) {
    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const {connectService, isPending} = useConnectServiceMutation(serviceName)
    const form = useRef<HTMLFormElement>(null)
    const authKeyHintId = useId()
    const [authKey, setAuthKey] = useState("")
    const [loginServerUrl, setLoginServerUrl] = useState("")
    const [hostname, setHostname] = useState("")

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()

        connectService({
            serviceName,
            authKey: authKey.trim() || undefined,
            loginServerUrl: loginServerUrl.trim() || undefined,
            hostname: hostname.trim() || undefined,
        })?.then(function handleConnected() {
            toaster.bake({title: "Tailscale sidecar started", description: serviceName, level: "Info"})
            if (form.current) CloseDialog()
        }).catch(toaster.catchGrpc)
    }

    return (
        <DialogShell title="Connect to Tailscale" onClose={CloseDialog} isFlush>
            <form ref={form} className={cls.TailscaleSidecarDialogContainer} onSubmit={handleSubmit}>
                <Input
                    className={cls.Field}
                    label="Auth key (optional)"
                    aria-label="Auth key (optional)"
                    type="password"
                    autoComplete="off"
                    aria-describedby={authKeyHintId}
                    value={authKey}
                    setValue={setAuthKey}
                    disabled={isPending}
                />
                <small className={cls.Hint} id={authKeyHintId}>Leave empty to use cluster Headscale</small>
                <Input
                    className={cls.Field}
                    label="Login server URL (optional)"
                    aria-label="Login server URL (optional)"
                    value={loginServerUrl}
                    setValue={setLoginServerUrl}
                    disabled={isPending}
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
