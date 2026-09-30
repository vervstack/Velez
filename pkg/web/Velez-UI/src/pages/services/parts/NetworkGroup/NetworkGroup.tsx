import cls from "@/pages/services/parts/NetworkGroup/NetworkGroup.module.css"
import type {DockerContainer} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import CreateServiceDialog from "@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx"
import ContainerCard from "@/pages/services/parts/ContainerCard/ContainerCard.tsx"

interface Props {
    root: DockerContainer
    members: DockerContainer[]
    onOpen: (id: string) => void
    onFilterByService: (serviceName: string) => void
}

export default function NetworkGroup({root, members, onOpen, onFilterByService}: Props) {
    const {OpenDialog} = useDialog()

    const isRegisterVisible = [root, ...members].some(function isUnregistered(container) {
        return !container.isRegistered
    })

    function handleRegister() {
        OpenDialog(<CreateServiceDialog container={root} networkMembers={members}/>)
    }

    return (
        <div className={cls.NetworkGroupContainer}>
            <div className={cls.Header}>
                <span className={cls.Caption}>Shared network</span>
                {isRegisterVisible && <Button variant="secondary" sm onClick={handleRegister}>Register</Button>}
            </div>
            <div className={cls.Cards}>
                <ContainerCard container={root} onOpen={onOpen} onFilterByService={onFilterByService} isRegisterHidden/>
                {members.map(function renderMember(member) {
                    return (
                        <ContainerCard
                            key={member.id}
                            container={member}
                            onOpen={onOpen}
                            onFilterByService={onFilterByService}
                            isRegisterHidden
                        />
                    )
                })}
            </div>
        </div>
    )
}
