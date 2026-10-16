import {SmerdStatus} from "@/app/api/velez"
import type {ServiceSidecar} from "@/app/api/velez"
import type {ServiceEnvironment, ServiceSidecarView} from "@/model/service_page/ServicePageModel"

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

function sidecarEnvironmentStatus(status?: SmerdStatus): ServiceEnvironment["status"] {
    switch (status) {
        case SmerdStatus.running:
            return "running"
        case SmerdStatus.restarting:
        case SmerdStatus.removing:
            return "degraded"
        case SmerdStatus.dead:
            return "failed"
        default:
            return "stopped"
    }
}

export function sidecarToEnvironment(sidecar: ServiceSidecarView): ServiceEnvironment {
    const isRunning = sidecar.status === SmerdStatus.running

    return {
        id: sidecar.containerId,
        label: "sidecar",
        status: sidecarEnvironmentStatus(sidecar.status),
        version: sidecar.imageName,
        deployedAgo: "",
        containerId: sidecar.containerId,
        containerName: sidecar.name,
        health: isRunning ? "healthy" : "unhealthy",
    }
}
