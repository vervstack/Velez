import type {LayoutItem} from "@/pages/services/processes/groupContainers.ts"
import ContainerCard from "@/pages/services/parts/ContainerCard/ContainerCard.tsx"
import NetworkGroup from "@/pages/services/parts/NetworkGroup/NetworkGroup.tsx"

interface Props {
    item: LayoutItem
    onOpen: (id: string) => void
    onFilterByService: (serviceName: string) => void
}

export default function ContainerItem({item, onOpen, onFilterByService}: Props) {
    if (item.kind === "network") {
        return (
            <NetworkGroup
                root={item.root}
                members={item.members}
                onOpen={onOpen}
                onFilterByService={onFilterByService}
            />
        )
    }

    return <ContainerCard container={item.container} onOpen={onOpen} onFilterByService={onFilterByService}/>
}
