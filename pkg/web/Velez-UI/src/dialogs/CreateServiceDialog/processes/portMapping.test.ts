import {describe, expect, it} from "vitest"

import {PortProtocol} from "@/app/api/velez"
import {
    describeRowOutcome,
    hasVolumes,
    parsePortMappingRows,
    PortMappingRow,
    portMappingRowsOf,
    publishedPortsOf,
    rowOutcomeOf,
} from "@/dialogs/CreateServiceDialog/processes/portMapping.ts"

function row(newHost: string, overrides: Partial<PortMappingRow> = {}): PortMappingRow {
    return {containerPort: 5432, protocol: PortProtocol.tcp, currentHost: 15432, newHost, ...overrides}
}

describe("portMapping", () => {
    it("lists only ports that are published on the host", () => {
        const container = {
            ports: [
                {servicePortNumber: 5432, exposedTo: 15432, protocol: PortProtocol.tcp},
                {servicePortNumber: 8080, protocol: PortProtocol.tcp},
            ],
        }

        expect(publishedPortsOf(container)).toEqual([container.ports[0]])
    })

    it("prefills each row with the current host port", () => {
        const container = {ports: [{servicePortNumber: 5432, exposedTo: 15432, protocol: PortProtocol.udp}]}

        expect(portMappingRowsOf(container)).toEqual([
            {containerPort: 5432, protocol: PortProtocol.udp, currentHost: 15432, newHost: "15432"},
        ])
    })

    it("defaults the protocol to tcp", () => {
        expect(portMappingRowsOf({ports: [{servicePortNumber: 80, exposedTo: 8080}]})[0].protocol)
            .toBe(PortProtocol.tcp)
    })

    it("classifies a row as kept, changed or unpublished", () => {
        expect(rowOutcomeOf(row("15432"))).toBe("kept")
        expect(rowOutcomeOf(row(" 15432 "))).toBe("kept")
        expect(rowOutcomeOf(row("15433"))).toBe("changed")
        expect(rowOutcomeOf(row("  "))).toBe("unpublished")
    })

    it("parses rows into ports and omits blank rows", () => {
        expect(parsePortMappingRows([row("15433"), row("", {containerPort: 80, currentHost: 8080})])).toEqual([
            {servicePortNumber: 5432, exposedTo: 15433, protocol: PortProtocol.tcp},
        ])
    })

    it("accepts no rows and all-blank rows", () => {
        expect(parsePortMappingRows([])).toEqual([])
        expect(parsePortMappingRows([row("")])).toEqual([])
    })

    it("rejects a non-numeric or out-of-range host port", () => {
        expect(parsePortMappingRows([row("abc")])).toBeNull()
        expect(parsePortMappingRows([row("0")])).toBeNull()
        expect(parsePortMappingRows([row("65536")])).toBeNull()
        expect(parsePortMappingRows([row("12.5")])).toBeNull()
        expect(parsePortMappingRows([row("65535")])).not.toBeNull()
        expect(parsePortMappingRows([row("1")])).not.toBeNull()
    })

    it("rejects two rows with the same new host port but allows several blank rows", () => {
        const other = {containerPort: 80, currentHost: 8080}

        expect(parsePortMappingRows([row("9000"), row("9000", other)])).toBeNull()
        expect(parsePortMappingRows([row(""), row("", other)])).toEqual([])
    })

    it("detects volumes and bind mounts", () => {
        expect(hasVolumes({mounts: [{type: "volume", destination: "/data"}]})).toBe(true)
        expect(hasVolumes({mounts: []})).toBe(false)
        expect(hasVolumes({})).toBe(false)
    })

    it("describes each outcome", () => {
        expect(describeRowOutcome(row("15432")))
            .toBe("Port 15432 is kept: the service is stopped completely (not deleted) until you finish onboarding.")
        expect(describeRowOutcome(row("15433")))
            .toBe("Port 15432 -> 15433: the old container is paused until you finish onboarding.")
        expect(describeRowOutcome(row(""))).toBe("Port 5432 will not be published.")
    })
})
