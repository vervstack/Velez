import {describe, expect, it} from "vitest"

import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {filterProvisioningBySearch, isServiceProvisioning} from "@/pages/services/processes/serviceProvisioning.ts"

function task(entityId: string, extra: Partial<ProvisioningTask> = {}): ProvisioningTask {
    return {taskId: "1", entityId, action: "create_service", status: ProvisioningTaskStatus.RUNNING, ...extra}
}

describe("isServiceProvisioning", () => {
    it("is true for an active task with the same name", () => {
        expect(isServiceProvisioning("api", [task("api")])).toBe(true)
    })

    it("is true for an active instance task whose bare name matches a prefixed service", () => {
        expect(isServiceProvisioning("pgaas_main", [task("main", {action: "drop_pg_instance"})])).toBe(true)
    })

    it("is true for an active register task whose entity id carries a uuid suffix", () => {
        expect(isServiceProvisioning("api", [task("api/8f2c", {action: "register_container"})])).toBe(true)
    })

    it("is false for a failed task", () => {
        expect(isServiceProvisioning("api", [task("api", {status: ProvisioningTaskStatus.FAILED})])).toBe(false)
    })

    it("is false for an unrelated name", () => {
        expect(isServiceProvisioning("web", [task("api")])).toBe(false)
    })
})

describe("filterProvisioningBySearch", () => {
    it("returns every task for a blank search", () => {
        const tasks = [task("api"), task("web")]
        expect(filterProvisioningBySearch(tasks, "  ")).toEqual(tasks)
    })

    it("keeps tasks whose name contains the query, ignoring case", () => {
        const tasks = [task("Api"), task("web")]
        expect(filterProvisioningBySearch(tasks, "AP")).toEqual([task("Api")])
    })
})
