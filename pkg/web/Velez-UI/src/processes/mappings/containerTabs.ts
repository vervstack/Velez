export type ContainerTab = "details" | "logs" | "inspect"

export const CONTAINER_TABS: Array<{id: ContainerTab, label: string}> = [
    {id: "details", label: "Details"},
    {id: "logs", label: "Logs"},
    {id: "inspect", label: "Inspect"},
]

export function parseContainerTab(value: string | null): ContainerTab {
    const found = CONTAINER_TABS.find(function matchesValue(tab) {
        return tab.id === value
    })
    return found ? found.id : "details"
}
