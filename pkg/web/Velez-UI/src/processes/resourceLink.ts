import {Routes} from "@/app/router/Routes"
import type {ServiceResource} from "@/model/service_page/ServicePageModel"

export function resourcePath(resource: Pick<ServiceResource, "linkedServiceName" | "containerId">): string | undefined {
    if (resource.linkedServiceName) {
        return Routes.Service + "/" + resource.linkedServiceName
    }
    if (resource.containerId) {
        return Routes.Container + "/" + resource.containerId
    }
    return undefined
}
