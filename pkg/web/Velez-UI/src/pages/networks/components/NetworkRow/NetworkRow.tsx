import cls from "@/pages/networks/components/NetworkRow/NetworkRow.module.css"
import type {Network, NetworkMember} from "@/app/api/velez/network_api.pb"
import {
    canAttachNetwork,
    canDeleteNetwork,
    getNetworkFlags,
    getNetworkKindLabel,
    hasDeleteCapability,
} from "@/processes/mappings/networks.ts"
import Button from "@/components/base/Button.tsx"
import NetworkMemberRow from "@/pages/networks/components/NetworkMemberRow/NetworkMemberRow.tsx"

interface Props {
    network: Network
    onAttach(network: Network): void
    onDelete(network: Network): void
    onDetach(network: Network, containerName: string): void
}

function renderFlag(flag: string) {
    return <span key={flag} className={cls.Flag}>{flag}</span>
}

export default function NetworkRow({network, onAttach, onDelete, onDetach}: Props) {
    const members = network.members ?? []
    const isAttachable = canAttachNetwork(network)

    function handleAttach() {
        onAttach(network)
    }

    function handleDelete() {
        onDelete(network)
    }

    function handleDetach(containerName: string) {
        onDetach(network, containerName)
    }

    function renderMember(member: NetworkMember) {
        return (
            <NetworkMemberRow
                key={member.containerId || member.containerName}
                member={member}
                canDetach={isAttachable}
                onDetach={handleDetach}
            />
        )
    }

    return (
        <div className={cls.NetworkRowContainer}>
            <div className={cls.Header}>
                <div className={cls.Meta}>
                    <span className={cls.Name}>{network.name}</span>
                    <span className={cls.Kind}>{getNetworkKindLabel(network)}</span>
                    {getNetworkFlags(network).map(renderFlag)}
                </div>
                <div className={cls.Actions}>
                    {isAttachable && <Button sm onClick={handleAttach}>Attach container</Button>}
                    {hasDeleteCapability(network) && (
                        <Button
                            sm
                            variant="danger"
                            onClick={handleDelete}
                            disabled={!canDeleteNetwork(network)}
                            tooltipContent={members.length > 0 ? "Detach all containers first" : undefined}
                        >
                            Delete
                        </Button>
                    )}
                </div>
            </div>
            {members.map(renderMember)}
        </div>
    )
}
