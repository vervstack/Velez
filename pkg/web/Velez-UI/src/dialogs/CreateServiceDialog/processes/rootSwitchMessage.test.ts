import {describe, expect, it} from "vitest"

import {rootSwitchMessage} from "@/dialogs/CreateServiceDialog/processes/rootSwitchMessage.ts"

const INFO = {oldRootName: "vpn", newRootName: "app", isClusterMode: false}

describe("rootSwitchMessage", () => {
    it("names both containers and states the downtime and the moved ports", () => {
        const message = rootSwitchMessage(INFO)

        expect(message).toContain("app becomes the network root and vpn rejoins its network as a sidecar")
        expect(message).toContain("briefly unavailable")
        expect(message).toContain("published ports, networks and aliases of vpn move to app")
    })

    it("mentions the recreation in cluster mode only", () => {
        expect(rootSwitchMessage(INFO)).not.toContain("cluster mode")
        expect(rootSwitchMessage({...INFO, isClusterMode: true})).toContain("In cluster mode")
    })
})
