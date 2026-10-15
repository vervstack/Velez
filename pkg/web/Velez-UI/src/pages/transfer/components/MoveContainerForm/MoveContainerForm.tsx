import {useState} from "react"

import cls from "@/pages/transfer/components/MoveContainerForm/MoveContainerForm.module.css"
import type {PlanTransferResponse} from "@/app/api/velez/transfer_api.pb"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {canStartTransfer} from "@/processes/mappings/transfer.ts"
import {PlanTransferMutation, StartTransferMutation} from "@/processes/queries/transfer.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import ContainerPicker from "@/pages/transfer/components/ContainerPicker/ContainerPicker.tsx"
import TransferPlanView from "@/pages/transfer/components/TransferPlanView/TransferPlanView.tsx"

export default function MoveContainerForm() {
    const [containerName, setContainerName] = useState("")
    const [address, setAddress] = useState("")
    const [token, setToken] = useState("")
    const [plan, setPlan] = useState<PlanTransferResponse | undefined>(undefined)
    const [selectedSidecars, setSelectedSidecars] = useState<string[]>([])
    const [isConsumersAcknowledged, setIsConsumersAcknowledged] = useState(false)

    const toaster = useToaster()
    const planTransfer = PlanTransferMutation()
    const startTransfer = StartTransferMutation()

    const isStartable = canStartTransfer({plan, containerName, address, token, isConsumersAcknowledged})
    const isCheckable = containerName !== "" && !planTransfer.isPending

    function resetPlan() {
        setPlan(undefined)
        setSelectedSidecars([])
        setIsConsumersAcknowledged(false)
    }

    function handleContainerChange(name: string) {
        setContainerName(name)
        resetPlan()
    }

    function handleCheck() {
        planTransfer.mutateAsync(containerName)
            .then(function handlePlanned(result) {
                setPlan(result)
                setSelectedSidecars((result.sidecars ?? []).map((sidecar) => sidecar.containerName ?? ""))
                setIsConsumersAcknowledged(false)
            })
            .catch(toaster.catchGrpc)
    }

    function handleSidecarToggle(name: string, isSelected: boolean) {
        setSelectedSidecars(isSelected
            ? [...selectedSidecars, name]
            : selectedSidecars.filter((selected) => selected !== name))
    }

    function handleStart() {
        startTransfer.mutateAsync({
            containerName,
            sidecarNames: selectedSidecars,
            address: address.trim(),
            token: token.trim(),
            isConsumersAcknowledged,
        })
            .then(function handleStarted() {
                toaster.bake({title: "Transfer started", description: containerName, level: "Info"})
                setToken("")
                resetPlan()
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <section className={cls.MoveContainerFormContainer}>
            <h2 className={cls.Title}>Move a container</h2>
            <div className={cls.FieldsWrapper}>
                <ContainerPicker value={containerName} onChange={handleContainerChange}/>
                <Input label="Target address" inputValue={address} onChange={setAddress} placeholder="host:53890"/>
                <Input label="Transfer token" inputValue={token} onChange={setToken}/>
            </div>
            <div className={cls.ActionsRow}>
                <Button onClick={handleCheck} disabled={!isCheckable}>
                    {planTransfer.isPending ? "Checking…" : "Check"}
                </Button>
            </div>
            {plan && (
                <TransferPlanView
                    plan={plan}
                    selectedSidecars={selectedSidecars}
                    isConsumersAcknowledged={isConsumersAcknowledged}
                    onSidecarToggle={handleSidecarToggle}
                    onConsumersAcknowledgedChange={setIsConsumersAcknowledged}
                />
            )}
            {plan && (
                <div className={cls.ActionsRow}>
                    <Button variant="primary" onClick={handleStart} disabled={!isStartable || startTransfer.isPending}>
                        {startTransfer.isPending ? "Starting…" : "Start transfer"}
                    </Button>
                </div>
            )}
        </section>
    )
}
