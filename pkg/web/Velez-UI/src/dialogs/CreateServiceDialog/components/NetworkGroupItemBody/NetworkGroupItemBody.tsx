import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/NetworkGroupItemBody/NetworkGroupItemBody.module.css"
import ContainerPreviewView, {ContainerPreviewState}
    from "@/dialogs/CreateServiceDialog/components/ContainerPreviewView/ContainerPreviewView.tsx"

interface Props {
    isOpen: boolean
    description: string
    preview: ContainerPreviewState
}

export default function NetworkGroupItemBody({isOpen, description, preview}: Props) {
    return (
        <div className={cn(cls.NetworkGroupItemBodyContainer, isOpen && cls.Open)}>
            <div className={cls.Inner}>
                <span className={cls.Description}>{description}</span>
                <ContainerPreviewView {...preview}/>
            </div>
        </div>
    )
}
