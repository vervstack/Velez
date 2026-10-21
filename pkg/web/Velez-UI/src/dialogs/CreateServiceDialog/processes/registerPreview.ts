import {
    ValueOrigin,
    type ContainerPreview,
    type PreviewRegisterContainerResponse,
    type PreviewValue,
} from "@/app/api/velez"

export type PreviewSectionKey = "networkMode" | "ports" | "env" | "labels" | "mounts"

export interface PreviewSection {
    key: PreviewSectionKey
    title: string
    values: PreviewValue[]
}

export function isAdded(value: PreviewValue): boolean {
    return value.origin === ValueOrigin.ADDED
}

function addedFirst(values: PreviewValue[]): PreviewValue[] {
    const added = values.filter(isAdded)
    const rest = values.filter((value) => !isAdded(value))
    return [...added, ...rest]
}

export function previewOfContainer(
    response: PreviewRegisterContainerResponse | undefined,
    containerId: string,
): ContainerPreview | undefined {
    return response?.containers?.find((preview) => preview.containerId === containerId)
}

export function previewSectionsOf(preview: ContainerPreview): PreviewSection[] {
    const networkMode = preview.networkMode?.key || preview.networkMode?.value ? [preview.networkMode] : []

    const sections: PreviewSection[] = [
        {key: "networkMode", title: "Network mode", values: networkMode as PreviewValue[]},
        {key: "ports", title: "Ports", values: addedFirst(preview.ports ?? [])},
        {key: "env", title: "Env", values: addedFirst(preview.env ?? [])},
        {key: "labels", title: "Labels", values: addedFirst(preview.labels ?? [])},
        {key: "mounts", title: "Mounts", values: addedFirst(preview.mounts ?? [])},
    ]

    return sections.filter((section) => section.values.length > 0)
}
