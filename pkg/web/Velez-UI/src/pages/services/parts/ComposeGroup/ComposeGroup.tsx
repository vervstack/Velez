import cls from "@/pages/services/parts/ComposeGroup/ComposeGroup.module.css"
import type {LayoutItem} from "@/pages/services/processes/groupContainers.ts"
import ContainerItem from "@/pages/services/parts/ContainerItem/ContainerItem.tsx"

interface Props {
    project: string
    items: LayoutItem[]
    onOpen: (id: string) => void
    onFilterByService: (serviceName: string) => void
}

export default function ComposeGroup({project, items, onOpen, onFilterByService}: Props) {
    return (
        <div className={cls.ComposeGroupContainer}>
            <span className={cls.Caption}>Compose project: {project}</span>
            <div className={cls.Items}>
                {items.map(function renderItem(item) {
                    const key = item.kind === "network" ? item.root.id : item.container.id
                    return <ContainerItem key={key} item={item} onOpen={onOpen} onFilterByService={onFilterByService}/>
                })}
            </div>
        </div>
    )
}
