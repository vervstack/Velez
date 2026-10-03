import cls from "@/pages/services/parts/ContainerSection/ContainerSection.module.css"
import type {LayoutEntry} from "@/pages/services/processes/groupContainers.ts"
import {containersOfEntry} from "@/pages/services/processes/partitionByRegistration.ts"
import ComposeGroup from "@/pages/services/parts/ComposeGroup/ComposeGroup.tsx"
import ContainerItem from "@/pages/services/parts/ContainerItem/ContainerItem.tsx"

interface Props {
    title: string
    entries: LayoutEntry[]
    onOpen: (id: string) => void
    onFilterByService: (serviceName: string) => void
}

function entryKey(entry: LayoutEntry): string {
    if (entry.kind === "compose") return "compose:" + entry.project
    return entry.kind === "network" ? entry.root.id ?? "" : entry.container.id ?? ""
}

export default function ContainerSection({title, entries, onOpen, onFilterByService}: Props) {
    const count = entries.reduce(function sum(total, entry) {
        return total + containersOfEntry(entry).length
    }, 0)

    return (
        <section className={cls.ContainerSectionContainer}>
            <h2 className={cls.Title}>{title} ({count})</h2>
            <div className={cls.Entries}>
                {entries.map(function renderEntry(entry) {
                    if (entry.kind === "compose") {
                        return (
                            <ComposeGroup
                                key={entryKey(entry)}
                                project={entry.project}
                                items={entry.items}
                                onOpen={onOpen}
                                onFilterByService={onFilterByService}
                            />
                        )
                    }
                    return (
                        <ContainerItem
                            key={entryKey(entry)}
                            item={entry}
                            onOpen={onOpen}
                            onFilterByService={onFilterByService}
                        />
                    )
                })}
            </div>
        </section>
    )
}
