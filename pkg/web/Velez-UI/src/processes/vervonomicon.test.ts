import {describe, it, expect} from "vitest"

import type {ServiceResource, ResourceReconciliationStatus} from "@/model/service_page/ServicePageModel"

import {
    buildVervonomiconTabs,
    findDefaultTab,
    isYamlFile,
    shouldHighlightYaml,
    mapResourceConnectionStatus,
    canCreateResource,
    mergeResourcesWithReconciliation,
} from "./vervonomicon"

describe("vervonomicon utilities", () => {
    describe("buildVervonomiconTabs", () => {
        it("creates tabs with matching id and label for each file path", () => {
            const paths = ["vervonomicon.yaml", "deployment.yaml", "prod/ingress.conf"]
            const tabs = buildVervonomiconTabs(paths)

            expect(tabs).toHaveLength(3)
            expect(tabs[0]).toEqual({
                id: "vervonomicon.yaml",
                label: "vervonomicon.yaml",
                filePath: "vervonomicon.yaml",
            })
            expect(tabs[2]).toEqual({
                id: "prod/ingress.conf",
                label: "prod/ingress.conf",
                filePath: "prod/ingress.conf",
            })
        })

        it("handles empty file list", () => {
            const tabs = buildVervonomiconTabs([])
            expect(tabs).toEqual([])
        })
    })

    describe("findDefaultTab", () => {
        it("returns vervonomicon.yaml when present", () => {
            const tabs = buildVervonomiconTabs(["prod/ingress.yaml", "vervonomicon.yaml", "deployment.yaml"])
            const defaultId = findDefaultTab(tabs)
            expect(defaultId).toBe("vervonomicon.yaml")
        })

        it("returns first tab when vervonomicon.yaml not present", () => {
            const tabs = buildVervonomiconTabs(["deployment.yaml", "ingress.yaml"])
            const defaultId = findDefaultTab(tabs)
            expect(defaultId).toBe("deployment.yaml")
        })

        it("returns empty string when no tabs", () => {
            const defaultId = findDefaultTab([])
            expect(defaultId).toBe("")
        })
    })

    describe("isYamlFile", () => {
        it("detects .yaml files", () => {
            expect(isYamlFile("vervonomicon.yaml")).toBe(true)
            expect(isYamlFile("deployment.yaml")).toBe(true)
        })

        it("detects .yml files", () => {
            expect(isYamlFile("config.yml")).toBe(true)
        })

        it("returns false for non-yaml files", () => {
            expect(isYamlFile("ingress.conf")).toBe(false)
            expect(isYamlFile("script.sh")).toBe(false)
        })

        it("handles paths with directories", () => {
            expect(isYamlFile("prod/ingress.yaml")).toBe(true)
            expect(isYamlFile("prod/ingress.conf")).toBe(false)
        })
    })

    describe("shouldHighlightYaml", () => {
        it("returns true for yaml files", () => {
            expect(shouldHighlightYaml("vervonomicon.yaml")).toBe(true)
            expect(shouldHighlightYaml("prod/deployment.yml")).toBe(true)
        })

        it("returns false for non-yaml files", () => {
            expect(shouldHighlightYaml("ingress.conf")).toBe(false)
        })
    })

    describe("mapResourceConnectionStatus", () => {
        it("maps each proto enum value to the matching union member", () => {
            expect(mapResourceConnectionStatus("RESOURCE_CONNECTION_STATUS_ALREADY_CONNECTED")).toBe("already_connected")
            expect(mapResourceConnectionStatus("RESOURCE_CONNECTION_STATUS_MUST_PROVISION")).toBe("must_provision")
        })

        it("maps unspecified and unknown/undefined values to unknown", () => {
            expect(mapResourceConnectionStatus("RESOURCE_CONNECTION_STATUS_UNSPECIFIED")).toBe("unknown")
            expect(mapResourceConnectionStatus("SOME_FUTURE_VALUE")).toBe("unknown")
            expect(mapResourceConnectionStatus(undefined)).toBe("unknown")
        })
    })

    describe("canCreateResource", () => {
        it("returns false for an already-connected resource", () => {
            expect(canCreateResource("already_connected")).toBe(false)
        })

        it("returns true for a resource that must be provisioned or has unknown status", () => {
            expect(canCreateResource("must_provision")).toBe(true)
            expect(canCreateResource("unknown")).toBe(true)
        })
    })

    describe("mergeResourcesWithReconciliation", () => {
        const getMeta = (resourceType: string) => ({icon: resourceType.slice(0, 2).toUpperCase(), color: "var(--fg-dim)"})

        function makeResource(overrides: Partial<ServiceResource>): ServiceResource {
            return {
                name: "db",
                type: "postgres",
                status: "healthy",
                icon: "Pg",
                color: "var(--info-color)",
                reconciliation: "unknown",
                ...overrides,
            }
        }

        it("overlays reconciliation status onto a matching bound resource", () => {
            const resources = [makeResource({name: "db"})]
            const reconciliation: ResourceReconciliationStatus[] = [
                {name: "db", resourceType: "postgres", status: "already_connected"},
            ]

            const merged = mergeResourcesWithReconciliation(resources, reconciliation, getMeta)

            expect(merged).toHaveLength(1)
            expect(merged[0].reconciliation).toBe("already_connected")
        })

        it("appends a placeholder for a declared resource that isn't bound yet", () => {
            const merged = mergeResourcesWithReconciliation(
                [],
                [{name: "cache", resourceType: "redis", status: "must_provision"}],
                getMeta,
            )

            expect(merged).toHaveLength(1)
            expect(merged[0]).toMatchObject({
                name: "cache",
                type: "redis",
                status: "unknown",
                reconciliation: "must_provision",
            })
        })

        it("does not duplicate a bound resource that is also declared", () => {
            const resources = [makeResource({name: "db"})]
            const reconciliation: ResourceReconciliationStatus[] = [
                {name: "db", resourceType: "postgres", status: "must_provision"},
            ]

            const merged = mergeResourcesWithReconciliation(resources, reconciliation, getMeta)

            expect(merged).toHaveLength(1)
        })

        it("leaves bound resources with no reconciliation entry as unknown", () => {
            const resources = [makeResource({name: "legacy", reconciliation: "unknown"})]

            const merged = mergeResourcesWithReconciliation(resources, [], getMeta)

            expect(merged).toEqual(resources)
        })
    })
})
