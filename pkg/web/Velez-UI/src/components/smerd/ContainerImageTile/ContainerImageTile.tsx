import cls from "@/components/smerd/ContainerImageTile/ContainerImageTile.module.css"

import {mapContainerStatusToDot} from "@/processes/mappings/containers"
import {imageMonogram} from "@/processes/mappings/serviceSidecars"
import type {SmerdStatus} from "@/app/api/velez"

import Button from "@/components/base/Button.tsx"
import StatusDot from "@/components/base/StatusDot.tsx"

interface Props {
    imageName: string
    label: string
    status?: SmerdStatus
    onOpen: () => void
}

export default function ContainerImageTile({imageName, label, status, onOpen}: Props) {
    return (
        <div className={cls.ContainerImageTileContainer}>
            <Button variant="ghost" borderless nopadding onClick={onOpen} tooltipContent={imageName}>
                <span className={cls.Monogram}>{imageMonogram(imageName)}</span>
                <span className={cls.Label}>{label}</span>
                <StatusDot status={mapContainerStatusToDot(status)}/>
            </Button>
        </div>
    )
}
