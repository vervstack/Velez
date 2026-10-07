import {describe, expect, it} from "vitest"

import {
    buildProxyPayload,
    buildRemoveProxyPayload,
    newBypassHostRow,
    toBypassHostRows,
} from "@/dialogs/ServiceProxyDialog/processes/buildProxyRequest.ts"

describe("buildProxyPayload", () => {
    it("returns null when the proxy url is blank", () => {
        expect(buildProxyPayload("  ", [])).toBeNull()
    })

    it("trims the url and drops blank hosts", () => {
        const rows = [newBypassHostRow(1, " postgres "), newBypassHostRow(2, " ")]

        expect(buildProxyPayload(" socks5://10.0.0.1:1080 ", rows)).toEqual({
            proxyUrl: "socks5://10.0.0.1:1080",
            proxyBypassHosts: ["postgres"],
        })
    })

    it("dedupes hosts keeping the first occurrence order", () => {
        const rows = toBypassHostRows(["docker", "postgres", " docker "])

        expect(buildProxyPayload("http://p:3128", rows)?.proxyBypassHosts).toEqual(["docker", "postgres"])
    })
})

describe("buildRemoveProxyPayload", () => {
    it("clears both the url and the hosts", () => {
        expect(buildRemoveProxyPayload()).toEqual({proxyUrl: "", proxyBypassHosts: []})
    })
})

describe("toBypassHostRows", () => {
    it("gives every host a unique id", () => {
        expect(toBypassHostRows(["a", "b"])).toEqual([{id: 1, host: "a"}, {id: 2, host: "b"}])
    })
})
