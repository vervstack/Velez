import {Checkbox} from "@vervstack/chures"

import cls from "@/pages/transfer/components/TransferPlanView/TransferPlanView.module.css"
import type {
    PlanTransferNetworkConsumer,
    PlanTransferResponse,
    PlanTransferSidecarCandidate,
} from "@/app/api/velez/transfer_api.pb"

interface Props {
    plan: PlanTransferResponse
    selectedSidecars: string[]
    isConsumersAcknowledged: boolean
    onSidecarToggle(name: string, isSelected: boolean): void
    onConsumersAcknowledgedChange(isAcknowledged: boolean): void
}

export default function TransferPlanView({
    plan,
    selectedSidecars,
    isConsumersAcknowledged,
    onSidecarToggle,
    onConsumersAcknowledgedChange,
}: Props) {
    const sidecars = plan.sidecars ?? []
    const consumers = plan.consumers ?? []
    const unsupportedMounts = plan.unsupportedMounts ?? []

    function renderSidecar(sidecar: PlanTransferSidecarCandidate) {
        const name = sidecar.containerName ?? ""

        function handleChange(isSelected: boolean) {
            onSidecarToggle(name, isSelected)
        }

        return (
            <div key={name} className={cls.Item}>
                <Checkbox label={name} checked={selectedSidecars.includes(name)} onChange={handleChange}/>
                <span className={cls.Meta}>{sidecar.imageName}</span>
            </div>
        )
    }

    function renderConsumer(consumer: PlanTransferNetworkConsumer) {
        const name = consumer.containerName ?? ""
        const network = consumer.networkName ?? ""

        return (
            <div key={`${name}/${network}`} className={cls.Item}>
                <span className={cls.Name}>{name}</span>
                <span className={cls.Meta}>{network}</span>
            </div>
        )
    }

    function renderMount(mount: string) {
        return <div key={mount} className={cls.Item}><span className={cls.Name}>{mount}</span></div>
    }

    return (
        <div className={cls.TransferPlanViewContainer}>
            {unsupportedMounts.length > 0 && (
                <div className={cls.ErrorBlock}>
                    <div className={cls.BlockTitle}>
                        Host mounts cannot be transferred. The transfer is blocked.
                    </div>
                    {unsupportedMounts.map(renderMount)}
                </div>
            )}
            {sidecars.length > 0 && (
                <div className={cls.Block}>
                    <div className={cls.BlockTitle}>Sidecars to move along</div>
                    <div className={cls.Hint}>
                        Unticked sidecars stay on the source and lose their link to the moved container.
                    </div>
                    {sidecars.map(renderSidecar)}
                </div>
            )}
            {consumers.length > 0 && (
                <div className={cls.WarnBlock}>
                    <div className={cls.BlockTitle}>
                        These containers use the same Docker network and will lose access.
                    </div>
                    {consumers.map(renderConsumer)}
                    <Checkbox
                        label="I understand"
                        checked={isConsumersAcknowledged}
                        onChange={onConsumersAcknowledgedChange}
                    />
                </div>
            )}
            {sidecars.length === 0 && consumers.length === 0 && unsupportedMounts.length === 0 && (
                <div className={cls.Hint}>No sidecars or network consumers. Ready to start.</div>
            )}
        </div>
    )
}
