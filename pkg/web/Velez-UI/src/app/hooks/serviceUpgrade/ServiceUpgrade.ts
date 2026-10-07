import {create} from "zustand"

import {TaskStatus, TaskStatusStatus} from "@/app/api/velez"
import {WatchServiceUpgradeStream} from "@/processes/api/tasks.ts"

export interface ServiceUpgrade {
    statusByService: Record<string, TaskStatus | undefined>
    watchingByService: Record<string, boolean | undefined>

    watch: (serviceName: string) => void
}

export const useServiceUpgrade = create<ServiceUpgrade>(
    (set, get) => ({
        statusByService: {},
        watchingByService: {},

        watch: (serviceName: string) => {
            if (get().watchingByService[serviceName]) {
                return
            }

            set((state) => ({
                statusByService: {...state.statusByService, [serviceName]: undefined},
                watchingByService: {...state.watchingByService, [serviceName]: true},
            }))

            WatchServiceUpgradeStream(serviceName, (status) => {
                set((state) => ({statusByService: {...state.statusByService, [serviceName]: status}}))
            })
                .catch(() => {
                    set((state) => {
                        const last = state.statusByService[serviceName]
                        const isTerminal = last?.status === TaskStatusStatus.DONE
                            || last?.status === TaskStatusStatus.FAILED
                        if (isTerminal) {
                            return state
                        }
                        return {statusByService: {...state.statusByService, [serviceName]: undefined}}
                    })
                })
                .finally(() => {
                    set((state) => ({watchingByService: {...state.watchingByService, [serviceName]: false}}))
                })
        },
    }))
