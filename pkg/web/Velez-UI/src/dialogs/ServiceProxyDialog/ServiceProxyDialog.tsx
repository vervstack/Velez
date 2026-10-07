import {useState} from "react"
import {useMutation} from "@tanstack/react-query"

import cls from "@/dialogs/ServiceProxyDialog/ServiceProxyDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {serviceService} from "@/processes/api/service.ts"
import Button from "@/components/base/Button.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import Input from "@/components/base/Input.tsx"
import BypassHostRow from "@/dialogs/ServiceProxyDialog/components/BypassHostRow/BypassHostRow.tsx"
import {
    BypassHostRow as BypassHostRowModel,
    buildProxyPayload,
    buildRemoveProxyPayload,
    newBypassHostRow,
    ProxyPayload,
    toBypassHostRows,
} from "@/dialogs/ServiceProxyDialog/processes/buildProxyRequest.ts"

const PROXY_HINT = "Velez recreates the container with HTTP_PROXY, HTTPS_PROXY, ALL_PROXY and NO_PROXY set " +
    "(upper- and lowercase). Only programs that read these variables use the proxy: curl, wget, git, Go programs, " +
    "Docker, Python requests and most CLIs. Node's built-in fetch, the JVM and raw TCP/UDP clients ignore them and " +
    "keep connecting directly. For a GitLab runner the proxy is also passed to CI jobs. Localhost and private " +
    "networks (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) always bypass the proxy — add internal hostnames such as " +
    "postgres or docker to Bypass hosts. Applying recreates the container, so running work (e.g. CI jobs) is " +
    "interrupted."

interface Props {
    serviceName: string
    currentProxyUrl: string
    currentBypassHosts: string[]
    onApplied: () => void
}

export default function ServiceProxyDialog({serviceName, currentProxyUrl, currentBypassHosts, onApplied}: Props) {
    const [proxyUrl, setProxyUrl] = useState(currentProxyUrl)
    const [rows, setRows] = useState<BypassHostRowModel[]>(() => toBypassHostRows(currentBypassHosts))
    const [nextRowId, setNextRowId] = useState(currentBypassHosts.length + 1)

    const {CloseDialog, LockClosing, UnlockClosing} = useDialog()
    const toaster = useToaster()
    const selectedEnvironment = useEnvironmentStore((state) => state.selectedEnvironment)

    const applyProxy = useMutation({
        mutationFn: (payload: ProxyPayload) => serviceService.setServiceProxy(
            serviceName,
            selectedEnvironment,
            payload.proxyUrl,
            payload.proxyBypassHosts,
        ),
    })

    const payload = buildProxyPayload(proxyUrl, rows)
    const isRemovable = currentProxyUrl !== ""

    function handleAddRow() {
        setRows([...rows, newBypassHostRow(nextRowId)])
        setNextRowId(nextRowId + 1)
    }

    function handleChangeRow(changed: BypassHostRowModel) {
        setRows(rows.map((row) => (row.id === changed.id ? changed : row)))
    }

    function handleRemoveRow(id: number) {
        setRows(rows.filter((row) => row.id !== id))
    }

    function submit(toSend: ProxyPayload, title: string) {
        LockClosing()
        applyProxy.mutateAsync(toSend)
            .then(function handleApplied() {
                toaster.bake({title, description: serviceName, level: "Info"})
                onApplied()
                UnlockClosing()
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
            .finally(UnlockClosing)
    }

    function handleApply() {
        if (!payload) return
        submit(payload, "Proxy applied")
    }

    function handleRemoveProxy() {
        submit(buildRemoveProxyPayload(), "Proxy removed")
    }

    function renderRow(row: BypassHostRowModel) {
        return <BypassHostRow key={row.id} row={row} onChange={handleChangeRow} onRemove={handleRemoveRow}/>
    }

    return (
        <div className={cls.ServiceProxyDialogContainer}>
            <DialogShell title={`Proxy for ${serviceName}`} onClose={CloseDialog}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Proxy URL"
                        inputValue={proxyUrl}
                        onChange={setProxyUrl}
                        placeholder="socks5://192.168.1.44:1080"
                        hint={PROXY_HINT}
                    />
                    {rows.map(renderRow)}
                    <Button sm onClick={handleAddRow}>Add host</Button>
                </div>
                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={applyProxy.isPending}>
                        Cancel
                    </Button>
                    {isRemovable && (
                        <Button variant="danger" onClick={handleRemoveProxy} disabled={applyProxy.isPending}>
                            Remove proxy
                        </Button>
                    )}
                    <Button variant="primary" onClick={handleApply} disabled={applyProxy.isPending || !payload}>
                        {applyProxy.isPending ? "Applying…" : "Apply"}
                    </Button>
                </div>
            </DialogShell>
        </div>
    )
}
