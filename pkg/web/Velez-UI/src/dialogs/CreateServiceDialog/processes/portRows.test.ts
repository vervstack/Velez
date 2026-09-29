import {describe, expect, it} from "vitest"

import {PortProtocol} from "@/app/api/velez"
import {describePort, parsePortRows, publishedPortsOf} from "@/dialogs/CreateServiceDialog/processes/portRows.ts"

describe("portRows", () => {
    it("lists only ports that are published on the host", () => {
        const container = {
            ports: [
                {servicePortNumber: 5432, exposedTo: 15432, protocol: PortProtocol.tcp},
                {servicePortNumber: 8080, protocol: PortProtocol.tcp},
            ],
        }

        expect(publishedPortsOf(container)).toEqual([container.ports[0]])
    })

    it("describes a port as host to container with its protocol", () => {
        expect(describePort({servicePortNumber: 5432, exposedTo: 15432, protocol: PortProtocol.udp}))
            .toBe("15432 → 5432/udp")
    })

    it("parses valid rows into tcp ports", () => {
        expect(parsePortRows([{containerPort: "5432", hostPort: " 15432 "}])).toEqual([
            {servicePortNumber: 5432, exposedTo: 15432, protocol: PortProtocol.tcp},
        ])
    })

    it("accepts no rows and ignores fully blank rows", () => {
        expect(parsePortRows([])).toEqual([])
        expect(parsePortRows([{containerPort: "", hostPort: " "}])).toEqual([])
    })

    it("rejects a half-filled, non-numeric or out-of-range row", () => {
        expect(parsePortRows([{containerPort: "5432", hostPort: ""}])).toBeNull()
        expect(parsePortRows([{containerPort: "abc", hostPort: "80"}])).toBeNull()
        expect(parsePortRows([{containerPort: "0", hostPort: "80"}])).toBeNull()
        expect(parsePortRows([{containerPort: "80", hostPort: "65536"}])).toBeNull()
    })
})
