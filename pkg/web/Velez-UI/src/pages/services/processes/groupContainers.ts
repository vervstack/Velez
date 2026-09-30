import type {DockerContainer} from "@/app/api/velez"

const COMPOSE_PROJECT_LABEL = "com.docker.compose.project"

export interface SingleItem {
    kind: "single"
    container: DockerContainer
}

export interface NetworkItem {
    kind: "network"
    root: DockerContainer
    members: DockerContainer[]
}

export type LayoutItem = SingleItem | NetworkItem

export interface ComposeEntry {
    kind: "compose"
    project: string
    items: LayoutItem[]
}

export type LayoutEntry = LayoutItem | ComposeEntry

interface Positioned<T> {
    position: number
    value: T
}

function composeProjectOf(container: DockerContainer): string {
    return container.labels?.[COMPOSE_PROJECT_LABEL] ?? ""
}

function resolveRoot(container: DockerContainer, byId: Map<string, DockerContainer>): DockerContainer {
    const visited = new Set<string>()
    let current = container
    while (current.networkOwnerContainerId) {
        visited.add(current.id ?? "")
        const owner = byId.get(current.networkOwnerContainerId)
        if (!owner) return current
        if (visited.has(owner.id ?? "")) return container
        current = owner
    }
    return current
}

function itemSize(item: LayoutItem): number {
    return item.kind === "network" ? 1 + item.members.length : 1
}

function itemProject(item: LayoutItem): string {
    return composeProjectOf(item.kind === "network" ? item.root : item.container)
}

function byPosition<T>(a: Positioned<T>, b: Positioned<T>): number {
    return a.position - b.position
}

function unwrap<T>(positioned: Positioned<T>): T {
    return positioned.value
}

export function groupContainers(containers: DockerContainer[]): LayoutEntry[] {
    const byId = new Map<string, DockerContainer>()
    const indexOf = new Map<DockerContainer, number>()
    containers.forEach(function index(container, position) {
        indexOf.set(container, position)
        if (container.id) byId.set(container.id, container)
    })

    const roots: DockerContainer[] = []
    const membersByRootId = new Map<string, DockerContainer[]>()
    containers.forEach(function place(container) {
        const root = resolveRoot(container, byId)
        if (root === container) {
            roots.push(container)
            return
        }
        const rootId = root.id ?? ""
        const members = membersByRootId.get(rootId) ?? []
        members.push(container)
        membersByRootId.set(rootId, members)
    })

    const items: Positioned<LayoutItem>[] = roots.map(function toItem(root) {
        const members = membersByRootId.get(root.id ?? "") ?? []
        const positions = [root, ...members].map(function positionOf(c) {
            return indexOf.get(c) ?? 0
        })
        const position = Math.min(...positions)
        const value: LayoutItem = members.length === 0
            ? {kind: "single", container: root}
            : {kind: "network", root, members}
        return {position, value}
    }).sort(byPosition)

    const entries: Positioned<LayoutEntry>[] = []
    const itemsByProject = new Map<string, Positioned<LayoutItem>[]>()
    items.forEach(function assign(item) {
        const project = itemProject(item.value)
        if (!project) {
            entries.push(item)
            return
        }
        const projectItems = itemsByProject.get(project) ?? []
        projectItems.push(item)
        itemsByProject.set(project, projectItems)
    })

    itemsByProject.forEach(function toEntries(projectItems, project) {
        const size = projectItems.reduce(function sum(total, item) {
            return total + itemSize(item.value)
        }, 0)
        if (size < 2) {
            entries.push(...projectItems)
            return
        }
        const value: ComposeEntry = {kind: "compose", project, items: projectItems.map(unwrap)}
        entries.push({position: projectItems[0].position, value})
    })

    return entries.sort(byPosition).map(unwrap)
}
