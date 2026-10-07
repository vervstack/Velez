import {describe, expect, it} from "vitest"

import {NetworkCapability} from "@/app/api/velez/network_api.pb"
import type {Network} from "@/app/api/velez/network_api.pb"
import {
    buildConnectContainerRequest,
    buildCreateNetworkRequest,
    canAttachNetwork,
    canDeleteNetwork,
    getAttachableContainerNames,
    getNetworkFlags,
    getNetworkKindLabel,
    hasDeleteCapability,
    parseAliases,
    shouldShowClusterBanner,
    sortNetworks,
} from "@/processes/mappings/networks.ts"

const DELETE = NetworkCapability.NETWORK_CAPABILITY_DELETE
const ATTACH = NetworkCapability.NETWORK_CAPABILITY_ATTACH

describe("shouldShowClusterBanner", () => {
    it("shows only when cluster mode is on and VCN is not connected", () => {
        expect(shouldShowClusterBanner({isClusterMode: true, isVcnConnected: false})).toBe(true)
        expect(shouldShowClusterBanner({isClusterMode: true})).toBe(true)
        expect(shouldShowClusterBanner({isClusterMode: true, isVcnConnected: true})).toBe(false)
        expect(shouldShowClusterBanner({isClusterMode: false, isVcnConnected: false})).toBe(false)
    })

    it("is hidden while the status is not loaded", () => {
        expect(shouldShowClusterBanner(undefined)).toBe(false)
        expect(shouldShowClusterBanner({})).toBe(false)
    })
})

describe("sortNetworks", () => {
    it("puts managed networks first, then sorts by name", () => {
        const networks: Network[] = [
            {name: "bridge"},
            {name: "zeta", isManaged: true},
            {name: "alpha", isManaged: true},
            {name: "host"},
        ]

        const names = sortNetworks(networks).map((n) => n.name)

        expect(names).toEqual(["alpha", "zeta", "bridge", "host"])
    })

    it("does not mutate the input", () => {
        const networks: Network[] = [{name: "b"}, {name: "a"}]

        sortNetworks(networks)

        expect(networks.map((n) => n.name)).toEqual(["b", "a"])
    })
})

describe("canDeleteNetwork", () => {
    it("requires the DELETE capability and no members", () => {
        expect(canDeleteNetwork({capabilities: [DELETE]})).toBe(true)
        expect(canDeleteNetwork({capabilities: [DELETE], members: []})).toBe(true)
        expect(canDeleteNetwork({capabilities: [DELETE], members: [{containerName: "app"}]})).toBe(false)
        expect(canDeleteNetwork({capabilities: [ATTACH]})).toBe(false)
        expect(canDeleteNetwork({})).toBe(false)
    })
})

describe("hasDeleteCapability", () => {
    it("ignores members", () => {
        expect(hasDeleteCapability({capabilities: [DELETE], members: [{containerName: "app"}]})).toBe(true)
        expect(hasDeleteCapability({capabilities: [ATTACH]})).toBe(false)
    })
})

describe("canAttachNetwork", () => {
    it("requires the ATTACH capability", () => {
        expect(canAttachNetwork({capabilities: [ATTACH]})).toBe(true)
        expect(canAttachNetwork({capabilities: [DELETE]})).toBe(false)
        expect(canAttachNetwork({})).toBe(false)
    })
})

describe("getNetworkFlags", () => {
    it("labels internal, isolated and subnet", () => {
        const flags = getNetworkFlags({isInternal: true, isIccEnabled: false, subnet: "172.20.0.0/16"})

        expect(flags).toEqual(["internal", "container isolation", "172.20.0.0/16"])
    })

    it("returns nothing for a plain network", () => {
        expect(getNetworkFlags({isInternal: false, isIccEnabled: true})).toEqual([])
        expect(getNetworkFlags({})).toEqual([])
    })
})

describe("getNetworkKindLabel", () => {
    it("distinguishes managed from foreign networks", () => {
        expect(getNetworkKindLabel({isManaged: true})).toBe("Velez")
        expect(getNetworkKindLabel({isManaged: false})).toBe("Docker (foreign)")
        expect(getNetworkKindLabel({})).toBe("Docker (foreign)")
    })
})

describe("parseAliases", () => {
    it("splits on commas, trims and drops blanks", () => {
        expect(parseAliases(" db , cache,, ")).toEqual(["db", "cache"])
        expect(parseAliases("")).toEqual([])
    })
})

describe("getAttachableContainerNames", () => {
    it("excludes members and sorts", () => {
        const network: Network = {members: [{containerName: "b"}]}

        expect(getAttachableContainerNames(["c", "b", "a"], network)).toEqual(["a", "c"])
    })
})

describe("buildCreateNetworkRequest", () => {
    it("returns null for a blank name", () => {
        const input = {name: "  ", environment: "PROD", isInternal: false, isIsolatingContainers: false}

        expect(buildCreateNetworkRequest(input)).toBeNull()
    })

    it("inverts isolation into isIccEnabled and trims the name", () => {
        const isolated = {name: " net ", environment: "PROD", isInternal: true, isIsolatingContainers: true}
        const open = {...isolated, isInternal: false, isIsolatingContainers: false}

        expect(buildCreateNetworkRequest(isolated)).toEqual(
            {name: "net", environment: "PROD", isInternal: true, isIccEnabled: false},
        )
        expect(buildCreateNetworkRequest(open)?.isIccEnabled).toBe(true)
    })
})

describe("buildConnectContainerRequest", () => {
    it("returns null without a container or a network id", () => {
        const base = {network: {id: "n1"}, containerName: "", aliasesText: "", environment: "PROD"}

        expect(buildConnectContainerRequest(base)).toBeNull()
        expect(buildConnectContainerRequest({...base, containerName: "app", network: {}})).toBeNull()
    })

    it("maps aliases from comma-separated text", () => {
        const input = {network: {id: "n1"}, containerName: "app", aliasesText: "a, b", environment: "PROD"}

        expect(buildConnectContainerRequest(input)).toEqual({
            networkId: "n1",
            containerName: "app",
            aliases: ["a", "b"],
            environment: "PROD",
        })
    })
})
