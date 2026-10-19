import {describe, expect, it} from "vitest"

import {Routes} from "@/app/router/Routes.ts"
import {aasRouteForLabels} from "@/processes/mappings/aasRoutes.ts"

describe("aasRouteForLabels", () => {
    it("returns the AAS page route for each instance type label", () => {
        expect(aasRouteForLabels(["service-runner-gitlab"])).toBe(Routes.Runners)
        expect(aasRouteForLabels(["service-pgaas"])).toBe(Routes.Postgres)
        expect(aasRouteForLabels(["service-s3"])).toBe(Routes.S3)
        expect(aasRouteForLabels(["service-dind"])).toBe(Routes.Dinds)
    })

    it("returns undefined when no label maps to an AAS page", () => {
        expect(aasRouteForLabels(["service-app"])).toBeUndefined()
        expect(aasRouteForLabels(undefined)).toBeUndefined()
    })
})
