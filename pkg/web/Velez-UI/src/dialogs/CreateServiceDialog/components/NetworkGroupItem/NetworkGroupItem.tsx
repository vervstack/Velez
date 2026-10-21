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
    canSwitch: boolean
}

interface Props {
    container: DockerContainer
    roles: ItemRoles
    description: string
    preview: ContainerPreviewState
    onMakeService(containerId: string): void
    onMakeRoot(containerId: string): void
}

export default function NetworkGroupItem({container, roles, description, preview, onMakeService, onMakeRoot}: Props) {
    const [isOpen, setIsOpen] = useState(false)
    const containerId = container.id ?? ""

    function handleToggle() {
        setIsOpen(!isOpen)
    }

    function handleMakeService() {
        onMakeService(containerId)
    }

    function handleMakeRoot() {
        onMakeRoot(containerId)
    }

    const actions = {
        canMakeService: roles.canSwitch && !roles.isService,
        canMakeRoot: roles.canSwitch && !roles.isRoot,
        onMakeService: handleMakeService,
        onMakeRoot: handleMakeRoot,
    }

    return (
        <div className={cn(cls.NetworkGroupItemContainer, isOpen && cls.Open)}>
            <ToggleButton variant="unstyled" className={cls.Header} aria-expanded={isOpen} onClick={handleToggle}>
                <span className={cls.Name}>{container.name}</span>
                <span className={cls.Image}>{container.imageName}</span>
                <RoleBadges isService={roles.isService} isRoot={roles.isRoot}/>
                <span className={cls.Chevron}>›</span>
            </ToggleButton>
            <NetworkGroupItemBody isOpen={isOpen} description={description} actions={actions} preview={preview}/>
        </div>
    )
}
