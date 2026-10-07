import cls from "@/pages/networks/components/NetworkMemberRow/NetworkMemberRow.module.css"
import type {NetworkMember} from "@/app/api/velez/network_api.pb"
import Button from "@/components/base/Button.tsx"

interface Props {
    member: NetworkMember
    canDetach: boolean
    onDetach(containerName: string): void
}

export default function NetworkMemberRow({member, canDetach, onDetach}: Props) {
    const containerName = member.containerName ?? ""
    const aliases = (member.aliases ?? []).join(", ")

    function handleDetach() {
        onDetach(containerName)
    }

    return (
        <div className={cls.NetworkMemberRowContainer}>
            <span className={cls.Name}>{containerName}</span>
            <span className={cls.Cell}>{aliases || "-"}</span>
            <span className={cls.Cell}>{member.ipAddress || "-"}</span>
            <div className={cls.Actions}>
                {canDetach && <Button sm variant="ghost" onClick={handleDetach}>Detach</Button>}
            </div>
        </div>
    )
}
