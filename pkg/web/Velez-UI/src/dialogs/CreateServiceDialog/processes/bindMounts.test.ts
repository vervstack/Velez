import {describe, expect, it} from "vitest"

import type {DockerContainer} from "@/app/api/velez"
import {
    bindMountsOf,
    defaultLinkVolumeName,
    resolveLinks,
} from "@/dialogs/CreateServiceDialog/processes/bindMounts.ts"

describe("bindMounts", () => {
    it("keeps only bind mounts that have a source and a destination", () => {
        const container: DockerContainer = {
            mounts: [
                {type: "bind", source: "/srv/data", destination: "/var/lib/data"},
                {type: "volume", source: "vol", destination: "/cache"},
                {type: "bind", source: "", destination: "/broken"},
            ],
        }

        expect(bindMountsOf(container)).toEqual([{source: "/srv/data", destination: "/var/lib/data"}])
    })

    it("returns no mounts for a container without any", () => {
        expect(bindMountsOf({})).toEqual([])
    })

    it("derives the volume name from the service and the trimmed target path", () => {
        expect(defaultLinkVolumeName("db", "/var/lib/postgresql/data")).toBe("db_var_lib_postgresql_data")
    })

    it("collapses runs of invalid characters into one underscore", () => {
        expect(defaultLinkVolumeName("db", "/data//my dir/")).toBe("db_data_my_dir")
    })

    it("keeps dots, dashes and underscores", () => {
        expect(defaultLinkVolumeName("my-app", "/etc/app.d/conf_1")).toBe("my-app_etc_app.d_conf_1")
    })

    it("prefers an override over the derived name", () => {
        const mounts = [
            {source: "/a", destination: "/x"},
            {source: "/b", destination: "/y"},
        ]

        const links = resolveLinks(mounts, "svc", {"/y": "custom"})

        expect(links).toEqual([
            {source: "/a", destination: "/x", volumeName: "svc_x"},
            {source: "/b", destination: "/y", volumeName: "custom"},
        ])
    })
})
