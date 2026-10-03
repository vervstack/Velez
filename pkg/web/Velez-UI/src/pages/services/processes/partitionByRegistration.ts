import type {DockerContainer} from "@/app/api/velez"
import type {ComposeEntry, LayoutEntry, LayoutItem} from "@/pages/services/processes/groupContainers.ts"

export type RegistrationFilter = "all" | "registered" | "unregistered"

export interface RegistrationPartition {
    registered: LayoutEntry[]
    unregistered: LayoutEntry[]
    awaitingEnd: LayoutEntry[]
}

type Bucket = keyof RegistrationPartition

const BUCKETS: Bucket[] = ["awaitingEnd", "registered", "unregistered"]

export function containersOfEntry(entry: LayoutEntry): DockerContainer[] {
    if (entry.kind === "single") return [entry.container]
    if (entry.kind === "network") return [entry.root, ...entry.members]
    return entry.items.flatMap(containersOfEntry)
}

export function itemsOfEntry(entry: LayoutEntry): LayoutItem[] {
    return entry.kind === "compose" ? entry.items : [entry]
}

function bucketOfItem(item: LayoutItem): Bucket {
    const containers = containersOfEntry(item)
    const isLeftover = containers.some(function hasReplacement(container) {
        return Boolean(container.replacedByContainerId)
    })
    if (isLeftover) return "awaitingEnd"
    const isRegistered = containers.every(function isRegistered(container) {
        return container.isRegistered === true
    })
    return isRegistered ? "registered" : "unregistered"
}

function containerCount(items: LayoutItem[]): number {
    return items.reduce(function sum(total, item) {
        return total + containersOfEntry(item).length
    }, 0)
}

function wrapItems(project: string, items: LayoutItem[]): LayoutEntry[] {
    if (items.length === 0) return []
    if (containerCount(items) < 2) return items
    const wrapped: ComposeEntry = {kind: "compose", project, items}
    return [wrapped]
}

function isBucketShown(bucket: Bucket, filter: RegistrationFilter): boolean {
    if (filter === "all") return true
    return bucket === "registered" ? filter === "registered" : filter === "unregistered"
}

export function partitionByRegistration(entries: LayoutEntry[], filter: RegistrationFilter): RegistrationPartition {
    const result: RegistrationPartition = {registered: [], unregistered: [], awaitingEnd: []}

    entries.forEach(function place(entry) {
        BUCKETS.forEach(function collect(bucket) {
            if (!isBucketShown(bucket, filter)) return
            const items = itemsOfEntry(entry).filter(function isInBucket(item) {
                return bucketOfItem(item) === bucket
            })
            if (entry.kind === "compose") {
                result[bucket].push(...wrapItems(entry.project, items))
                return
            }
            result[bucket].push(...items)
        })
    })

    return result
}
