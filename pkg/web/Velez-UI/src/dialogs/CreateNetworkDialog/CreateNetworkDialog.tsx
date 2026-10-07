import {useState} from "react"
import {Checkbox} from "@vervstack/chures"
import {useNavigate} from "react-router-dom"

import cls from "@/dialogs/CreateNetworkDialog/CreateNetworkDialog.module.css"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {Routes} from "@/app/router/Routes.ts"
import {buildCreateNetworkRequest, shouldShowClusterBanner} from "@/processes/mappings/networks.ts"
import {CreateNetworkMutation, useNetworkStatusQuery} from "@/processes/queries/networks.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import NetworkClusterBanner from "@/components/NetworkClusterBanner/NetworkClusterBanner.tsx"

export default function CreateNetworkDialog() {
    const [name, setName] = useState("")
    const [isInternal, setIsInternal] = useState(false)
    const [isIsolatingContainers, setIsIsolatingContainers] = useState(false)

    const navigate = useNavigate()
    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const environment = useEnvironmentStore((state) => state.selectedEnvironment)
    const statusQuery = useNetworkStatusQuery()
    const createNetwork = CreateNetworkMutation()

    const req = buildCreateNetworkRequest({name, environment, isInternal, isIsolatingContainers})

    function handleOpenVcn() {
        CloseDialog()
        navigate(Routes.VCN)
    }

    function handleCreate() {
        if (!req) return

        createNetwork.mutateAsync(req)
            .then(function handleCreated() {
                toaster.bake({title: "Network created", description: req.name ?? "", level: "Info"})
                CloseDialog()
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.CreateNetworkDialogContainer}>
            <DialogShell title="Create network" onClose={CloseDialog}>
                {shouldShowClusterBanner(statusQuery.data) && <NetworkClusterBanner onOpenVcn={handleOpenVcn}/>}
                <div className={cls.FieldsWrapper}>
                    <Input label="Name" inputValue={name} onChange={setName} disabled={createNetwork.isPending}/>
                    <Checkbox
                        label="Internal (no access outside the network)"
                        checked={isInternal}
                        onChange={setIsInternal}
                        disabled={createNetwork.isPending}
                    />
                    <Checkbox
                        label="Isolate containers from each other (disable container-to-container traffic)"
                        checked={isIsolatingContainers}
                        onChange={setIsIsolatingContainers}
                        disabled={createNetwork.isPending}
                    />
                </div>
                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={createNetwork.isPending}>
                        Cancel
                    </Button>
                    <Button variant="primary" onClick={handleCreate} disabled={createNetwork.isPending || !req}>
                        {createNetwork.isPending ? "Creating…" : "Create"}
                    </Button>
                </div>
            </DialogShell>
        </div>
    )
}
