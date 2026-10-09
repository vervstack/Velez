import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {isProvisioningActive, provisioningName} from "@/processes/mappings/provisioning.ts"

// Instance names of AsAService kinds are stored in Docker with these prefixes;
// the matching task keeps the bare name as its entity id.
const SERVICE_NAME_PREFIXES = ["", "pgaas_", "cr_", "s3_", "gitlab_runner_", "github_runner_"]

export function isServiceProvisioning(serviceName: string, tasks: ProvisioningTask[]): boolean {
    return tasks.some((task) => {
        if (!isProvisioningActive(task)) return false
        const name = provisioningName(task)
        return SERVICE_NAME_PREFIXES.some((prefix) => prefix + name === serviceName)
    })
}

export function filterProvisioningBySearch(tasks: ProvisioningTask[], search: string): ProvisioningTask[] {
    const query = search.trim().toLowerCase()
    if (!query) return tasks
    return tasks.filter((task) => provisioningName(task).toLowerCase().includes(query))
}
