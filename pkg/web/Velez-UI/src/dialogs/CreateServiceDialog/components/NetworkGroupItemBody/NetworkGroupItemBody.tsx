import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/NetworkGroupItemBody/NetworkGroupItemBody.module.css"
import ContainerPreviewView, {ContainerPreviewState}
    from "@/dialogs/CreateServiceDialog/components/ContainerPreviewView/ContainerPreviewView.tsx"
import NetworkGroupItemActions, {ItemActionsProps}
    from "@/dialogs/CreateServiceDialog/components/NetworkGroupItemActions/NetworkGroupItemActions.tsx"

interface Props {
    isOpen: boolean
    description: string
    actions: ItemActionsProps
    preview: ContainerPreviewState
}

export default function NetworkGroupItemBody({isOpen, description, actions, preview}: Props) {
    return (
        <div className={cn(cls.NetworkGroupItemBodyContainer, isOpen && cls.Open)}>
            <div className={cls.Inner}>
                <span className={cls.Description}>{description}</span>
                <NetworkGroupItemActions {...actions}/>
                <ContainerPreviewView {...preview}/>
            </div>
        </div>
    )
}
