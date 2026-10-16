import cn from "classnames"

import type {SmerdStatus} from "@/app/api/velez"
import cls from "@/pages/container/components/ContainerStatusBadge/ContainerStatusBadge.module.css"
import {containerStatusTone} from "@/processes/mappings/containers.ts"

interface Props {
    status?: SmerdStatus
    isUnregistered?: boolean
}

export default function ContainerStatusBadge({status, isUnregistered}: Props) {
    if (isUnregistered) {
        return <span className={cn(cls.ContainerStatusBadgeContainer, cls.Unknown)}>Unregistered</span>
    }
    if (!status) return null

    const tone = containerStatusTone(status)
    return (
        <span className={cn(cls.ContainerStatusBadgeContainer, {
            [cls.Running]: tone === "running",
            [cls.Error]: tone === "error",
            [cls.Unknown]: tone === "unknown",
        })}>
            {status}
        </span>
    )
}
