import {useQueryClient} from "@tanstack/react-query"

import type {FinalStepProps} from "@/dialogs/StepsDialog/Step.ts"
import {HAPROXY_INSTANCES_QUERY_KEY, useCreateHaproxyInstanceMutation} from "@/processes/queries/haproxy.ts"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"
import {
    createRequest,
    type HaproxyWizardContext,
} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"

export default function CreateHaproxyProgress({context, onClose}: FinalStepProps<HaproxyWizardContext>) {
    const queryClient = useQueryClient()
    const createInstance = useCreateHaproxyInstanceMutation()
    const request = createRequest(context)

    function start() {
        return createInstance.mutateAsync(request)
    }

    function refresh() {
        void queryClient.invalidateQueries({queryKey: HAPROXY_INSTANCES_QUERY_KEY})
        void queryClient.invalidateQueries({queryKey: ["services"]})
    }

    return (
        <TaskProgressScreen title="Creating HAProxy instance" metaLine={request.name ?? ""}
                            start={start}
                            onSuccess={refresh} onClose={onClose}/>
    )
}
