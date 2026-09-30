import cls from "@/dialogs/CreateServiceDialog/components/NetworkMembersNotice/NetworkMembersNotice.module.css"
import type {DockerContainer} from "@/app/api/velez"

interface Props {
    members: DockerContainer[]
}

export default function NetworkMembersNotice({members}: Props) {
    return (
        <div className={cls.NetworkMembersNoticeContainer} role="note">
            <span className={cls.Title}>Also migrating with this container:</span>
            <ul className={cls.List}>
                {members.map(function renderMember(member) {
                    return <li key={member.id}>{member.name} ({member.imageName})</li>
                })}
            </ul>
        </div>
    )
}
