import {useMutation, useQueryClient} from "@tanstack/react-query"

import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {tasksService} from "@/processes/api/tasks.ts"

export function DismissTaskMutation(queryKey: readonly unknown[]) {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (task: ProvisioningTask) => tasksService.dismissTask(task.entityId ?? "", task.action ?? ""),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey})
        },
    })
}
