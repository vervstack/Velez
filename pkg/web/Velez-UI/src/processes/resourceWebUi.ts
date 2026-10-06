import type {ServiceResource} from "@/model/service_page/ServicePageModel"

export function resourceWebUiUrl(
    resource: Pick<ServiceResource, "webUiPort" | "webUiHost">,
    location: Pick<Location, "protocol" | "hostname">
): string | undefined {
    if (!resource.webUiPort || resource.webUiPort === 0) {
        return undefined
    }
    const host = resource.webUiHost || location.hostname
    return `${location.protocol}//${host}:${resource.webUiPort}`
}
