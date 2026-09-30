import type {ServiceSidecar} from "@/app/api/velez"
import type {ServiceSidecarView} from "@/model/service_page/ServicePageModel"

export function toServiceSidecars(sidecars?: ServiceSidecar[]): ServiceSidecarView[] {
    return (sidecars ?? [])
        .filter(sidecar => !!sidecar.containerId)
        .map(sidecar => ({
            containerId: sidecar.containerId ?? "",
            name: sidecar.containerName || sidecar.containerId || "",
            imageName: sidecar.imageName ?? "",
            status: sidecar.status,
        }))
}

export function imageMonogram(imageName?: string): string {
    const lastSegment = (imageName ?? "").split("/").pop() ?? ""
    const repository = lastSegment.split(":")[0].split("@")[0]
    const first = repository.charAt(0)
    return first ? first.toUpperCase() : "?"
}
