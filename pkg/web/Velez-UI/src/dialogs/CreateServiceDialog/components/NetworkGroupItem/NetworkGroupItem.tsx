import {useState} from "react"
import {Button as ToggleButton} from "@vervstack/chures"
import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/NetworkGroupItem/NetworkGroupItem.module.css"
import type {DockerContainer} from "@/app/api/velez"
import type {ContainerPreviewState}
    from "@/dialogs/CreateServiceDialog/components/ContainerPreviewView/ContainerPreviewView.tsx"
import NetworkGroupItemBody
    from "@/dialogs/CreateServiceDialog/components/NetworkGroupItemBody/NetworkGroupItemBody.tsx"
import RoleBadges from "@/dialogs/CreateServiceDialog/components/RoleBadges/RoleBadges.tsx"

export interface ItemRoles {
    isRoot: boolean
    isService: boolean
}

interface Props {
    container: DockerContainer
    roles: ItemRoles
    description: string
    preview: ContainerPreviewState
}

export default function NetworkGroupItem({container, roles, description, preview}: Props) {
    const [isOpen, setIsOpen] = useState(false)

    function handleToggle() {
        setIsOpen(!isOpen)
    }

    return (
        <div className={cn(cls.NetworkGroupItemContainer, isOpen && cls.Open)}>
            <ToggleButton variant="unstyled" className={cls.Header} aria-expanded={isOpen} onClick={handleToggle}>
                <span className={cls.Name}>{container.name}</span>
                <span className={cls.Image}>{container.imageName}</span>
                <RoleBadges isService={roles.isService} isRoot={roles.isRoot}/>
                <span className={cls.Chevron}>›</span>
            </ToggleButton>
            <NetworkGroupItemBody isOpen={isOpen} description={description} preview={preview}/>
        </div>
    )
}
