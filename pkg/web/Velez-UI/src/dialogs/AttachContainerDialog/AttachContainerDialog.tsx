import {useMemo, useState} from "react"
import {Dropdown, DropdownOption} from "@vervstack/chures"
import {useNavigate} from "react-router-dom"

import cls from "@/dialogs/AttachContainerDialog/AttachContainerDialog.module.css"
import type {Network} from "@/app/api/velez/network_api.pb"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {Routes} from "@/app/router/Routes.ts"
import {
    buildConnectContainerRequest,
    getAttachableContainerNames,
    shouldShowClusterBanner,
} from "@/processes/mappings/networks.ts"
import {useListContainersQuery} from "@/processes/queries/containers.ts"
import {ConnectContainerMutation, useNetworkStatusQuery} from "@/processes/queries/networks.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import NetworkClusterBanner from "@/components/NetworkClusterBanner/NetworkClusterBanner.tsx"

interface Props {
    network: Network
}

export default function AttachContainerDialog({network}: Props) {
    const [containerName, setContainerName] = useState("")
    const [aliasesText, setAliasesText] = useState("")

    const navigate = useNavigate()
    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const environment = useEnvironmentStore((state) => state.selectedEnvironment)
    const statusQuery = useNetworkStatusQuery()
    const containersQuery = useListContainersQuery()
    const connectContainer = ConnectContainerMutation()

    const options: DropdownOption[] = useMemo(() => {
        const names = (containersQuery.data?.containers ?? []).map((c) => c.name ?? "").filter(Boolean)
        return getAttachableContainerNames(names, network).map((name) => ({id: name, name}))
    }, [containersQuery.data, network])

    const req = buildConnectContainerRequest({network, containerName, aliasesText, environment})

    function handleContainerChange(ids: string[]) {
        setContainerName(ids[0] ?? "")
    }

    function handleOpenVcn() {
        CloseDialog()
        navigate(Routes.VCN)
    }

    function handleAttach() {
        if (!req) return

        connectContainer.mutateAsync(req)
            .then(function handleAttached() {
                toaster.bake({title: "Container attached", description: req.containerName ?? "", level: "Info"})
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.AttachContainerDialogContainer}>
            <DialogShell title={`Attach container to ${network.name}`} onClose={CloseDialog}>
                {shouldShowClusterBanner(statusQuery.data) && <NetworkClusterBanner onOpenVcn={handleOpenVcn}/>}
                <div className={cls.FieldsWrapper}>
                    <Dropdown
                        label="Container"
                        placeholder="Select container"
                        options={options}
                        value={containerName ? [containerName] : []}
                        onChange={handleContainerChange}
                        isLoading={containersQuery.isLoading}
                        portal
                    />
                    <Input
                        label="Aliases (comma-separated, optional)"
                        inputValue={aliasesText}
                        onChange={setAliasesText}
                        disabled={connectContainer.isPending}
                    />
                </div>
                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={connectContainer.isPending}>
                        Cancel
                    </Button>
                    <Button variant="primary" onClick={handleAttach} disabled={connectContainer.isPending || !req}>
                        {connectContainer.isPending ? "Attaching…" : "Attach"}
                    </Button>
                </div>
            </DialogShell>
        </div>
    )
}
