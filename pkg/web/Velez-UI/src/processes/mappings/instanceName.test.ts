import {describe, expect, it} from "vitest"

import {validateInstanceName} from "@/processes/mappings/instanceName.ts"

const MESSAGE =
    "Instance name must be 2-32 characters: lowercase letters, digits, - and _, starting with a letter or digit"

describe("validateInstanceName", () => {
    it.each(["ft", "a1", "my-db_1", "a".repeat(32), "", "  ft  "])("accepts %j", (name) => {
        expect(validateInstanceName(name)).toBeUndefined()
    })

    it.each(["a".repeat(33), "a", "Foo", "-x", "_x", "x y", "x.y"])("rejects %j with the message", (name) => {
        expect(validateInstanceName(name)).toBe(MESSAGE)
    })
})
