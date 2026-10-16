import cn from "classnames"

import cls from "@/pages/container/components/ContainerLogDetailDrawer/ContainerLogDetailDrawer.module.css"
import {LogLine} from "@/processes/mappings/containerLogs.ts"
import ContainerLogDetailPanel from
    "@/pages/container/components/ContainerLogDetailPanel/ContainerLogDetailPanel.tsx"

interface Props {
    line: LogLine | null
    query: string
    isOpen: boolean
    onClose: () => void
}

export default function ContainerLogDetailDrawer({line, query, isOpen, onClose}: Props) {
    return (
        <div className={cn(cls.ContainerLogDetailDrawerContainer, {[cls.Open]: isOpen})} aria-hidden={!isOpen}>
            <div className={cls.Inner}>
                <ContainerLogDetailPanel line={line} query={query} onClose={onClose}/>
            </div>
        </div>
    )
}
