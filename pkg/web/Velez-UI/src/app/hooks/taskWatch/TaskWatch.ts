import {create} from "zustand"

import {TaskStatus, TaskStatusStatus} from "@/app/api/velez"
import {WatchTaskStream} from "@/processes/api/tasks.ts"
import {taskKey} from "@/processes/taskKey.ts"

export interface TaskWatch {
    statusByTask: Record<string, TaskStatus | undefined>
    watchingByTask: Record<string, boolean | undefined>

    watch: (entityId: string, action: string) => void
}

export const useTaskWatch = create<TaskWatch>(
    (set, get) => ({
        statusByTask: {},
        watchingByTask: {},

        watch: (entityId: string, action: string) => {
            const key = taskKey(entityId, action)
            if (get().watchingByTask[key]) {
                return
            }

            set((state) => ({
                statusByTask: {...state.statusByTask, [key]: undefined},
                watchingByTask: {...state.watchingByTask, [key]: true},
            }))

            WatchTaskStream({entityId, action}, (status) => {
                set((state) => ({statusByTask: {...state.statusByTask, [key]: status}}))
            })
                .catch(() => {
                    set((state) => {
                        const last = state.statusByTask[key]
                        const isTerminal = last?.status === TaskStatusStatus.DONE
                            || last?.status === TaskStatusStatus.FAILED
                        if (isTerminal) {
                            return state
                        }
                        return {statusByTask: {...state.statusByTask, [key]: undefined}}
                    })
                })
                .finally(() => {
                    set((state) => ({watchingByTask: {...state.watchingByTask, [key]: false}}))
                })
        },
    }))
