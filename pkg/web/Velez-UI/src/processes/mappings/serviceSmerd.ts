import type {Smerd} from "@/app/api/velez"

// Must match labels.VervServiceLabel in internal/domain/labels/verv_labels.go.
const VERV_SERVICE_LABEL = "VERV_SERVICE"

export function isSmerdOfService(smerd: Smerd, serviceName: string): boolean {
    if (!serviceName) {
        return false
    }

    const labelledService = smerd.labels?.[VERV_SERVICE_LABEL]
    if (labelledService) {
        return labelledService === serviceName
    }

    const containerName = smerd.name
    if (!containerName) {
        return false
    }

    return containerName === serviceName || containerName.startsWith(serviceName + "-")
}

export function findSmerdOfService(smerds: Smerd[], serviceName: string): Smerd | undefined {
    return smerds.find(smerd => isSmerdOfService(smerd, serviceName))
}
