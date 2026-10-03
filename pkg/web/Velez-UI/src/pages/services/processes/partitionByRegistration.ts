import type {DockerContainer} from "@/app/api/velez"
import type {LayoutEntry} from "@/pages/services/processes/groupContainers.ts"

export type RegistrationFilter = "all" | "registered" | "unregistered"

export interface RegistrationPartition {
    registered: LayoutEntry[]
    unregistered: LayoutEntry[]
}

export function containersOfEntry(entry: LayoutEntry): DockerContainer[] {
    if (entry.kind === "single") return [entry.container]
    if (entry.kind === "network") return [entry.root, ...entry.members]
    return entry.items.flatMap(containersOfEntry)
}

function isEntryRegistered(entry: LayoutEntry): boolean {
    return containersOfEntry(entry).every(function isRegistered(container) {
        return container.isRegistered === true
    })
}

export function partitionByRegistration(entries: LayoutEntry[], filter: RegistrationFilter): RegistrationPartition {
    const registered: LayoutEntry[] = []
    const unregistered: LayoutEntry[] = []

    entries.forEach(function place(entry) {
        if (isEntryRegistered(entry)) {
            if (filter !== "unregistered") registered.push(entry)
            return
        }
        if (filter !== "registered") unregistered.push(entry)
    })

    return {registered, unregistered}
}
