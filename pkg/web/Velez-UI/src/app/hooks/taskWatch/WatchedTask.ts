import {useEffect} from "react"

import {TaskStatus} from "@/app/api/velez"
import {useTaskWatch} from "@/app/hooks/taskWatch/TaskWatch.ts"
import {taskKey} from "@/processes/taskKey.ts"

export function useWatchedTask(entityId: string, action: string): TaskStatus | undefined {
    const watch = useTaskWatch((state) => state.watch)
    const status = useTaskWatch((state) => state.statusByTask[taskKey(entityId, action)])

    useEffect(() => {
        watch(entityId, action)
    }, [entityId, action, watch])

    return status
}
