import {act, renderHook} from "@testing-library/react"
import {QueryClient, QueryClientProvider} from "@tanstack/react-query"
import {afterEach, describe, expect, it, vi} from "vitest"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useInvalidateOnDialogClose} from "@/app/hooks/InvalidateOnDialogClose.ts"

const queryKey = ["services"]

function renderInvalidateHook() {
    const queryClient = new QueryClient()
    const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue(undefined)

    function wrapper({children}: { children: React.ReactNode }) {
        return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    }

    renderHook(() => useInvalidateOnDialogClose(queryKey), {wrapper})
    return {invalidateQueries}
}

describe("useInvalidateOnDialogClose", () => {
    afterEach(() => {
        act(() => useDialog.setState({children: null}))
    })

    it("does not invalidate while the dialog is opening", () => {
        const {invalidateQueries} = renderInvalidateHook()

        act(() => useDialog.getState().OpenDialog(<span>dialog</span>))

        expect(invalidateQueries).not.toHaveBeenCalled()
    })

    it("invalidates the query when the dialog closes", () => {
        const {invalidateQueries} = renderInvalidateHook()

        act(() => useDialog.getState().OpenDialog(<span>dialog</span>))
        act(() => useDialog.getState().CloseDialog())

        expect(invalidateQueries).toHaveBeenCalledWith({queryKey})
    })
})
