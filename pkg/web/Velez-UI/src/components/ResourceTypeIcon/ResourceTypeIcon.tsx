import type {ComponentType} from "react"

import cls from "@/components/ResourceTypeIcon/ResourceTypeIcon.module.css"

import DindIcon from "@/components/base/icons/DindIcon.tsx"
import PostgresIcon from "@/components/base/icons/PostgresIcon.tsx"
import ProxyIcon from "@/components/base/icons/ProxyIcon.tsx"
import RegistryIcon from "@/components/base/icons/RegistryIcon.tsx"
import S3Icon from "@/components/base/icons/S3Icon.tsx"

interface Props {
    type: string
    fallback: string
}

const ICONS: Record<string, ComponentType<{className?: string}>> = {
    dind: DindIcon,
    postgres: PostgresIcon,
    s3: S3Icon,
    registry: RegistryIcon,
    proxy: ProxyIcon,
}

export default function ResourceTypeIcon({type, fallback}: Props) {
    const Icon = ICONS[type.toLowerCase()]

    return (
        <span className={cls.ResourceTypeIconContainer}>
            {Icon ? <Icon className={cls.Icon}/> : <span>{fallback}</span>}
        </span>
    )
}
