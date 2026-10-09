import {useEffect, useRef} from "react"
import {useQueryClient} from "@tanstack/react-query"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"

export function useInvalidateOnDialogClose(queryKey: readonly unknown[]) {
    const queryClient = useQueryClient()
    const isDialogOpen = useDialog((state) => state.children !== null)
    const wasDialogOpen = useRef(isDialogOpen)

    useEffect(() => {
        if (wasDialogOpen.current && !isDialogOpen) {
            queryClient.invalidateQueries({queryKey})
        }
        wasDialogOpen.current = isDialogOpen
    }, [isDialogOpen])
}
