import {afterEach, beforeEach, describe, expect, it, vi} from "vitest"
import {act, renderHook} from "@testing-library/react"

import {useDebouncedValue} from "@/app/hooks/useDebouncedValue.ts"

beforeEach(() => {
    vi.useFakeTimers()
})

afterEach(() => {
    vi.useRealTimers()
})

describe("useDebouncedValue", () => {
    it("keeps the previous value until the delay has passed", () => {
        const {result, rerender} = renderHook(({value}) => useDebouncedValue(value, 300), {
            initialProps: {value: "a"},
        })

        rerender({value: "ab"})
        expect(result.current).toBe("a")

        act(() => {
            vi.advanceTimersByTime(300)
        })
        expect(result.current).toBe("ab")
    })

    it("restarts the delay when the value changes again", () => {
        const {result, rerender} = renderHook(({value}) => useDebouncedValue(value, 300), {
            initialProps: {value: "a"},
        })

        rerender({value: "ab"})
        act(() => {
            vi.advanceTimersByTime(200)
        })
        rerender({value: "abc"})
        act(() => {
            vi.advanceTimersByTime(200)
        })
        expect(result.current).toBe("a")

        act(() => {
            vi.advanceTimersByTime(100)
        })
        expect(result.current).toBe("abc")
    })
})
