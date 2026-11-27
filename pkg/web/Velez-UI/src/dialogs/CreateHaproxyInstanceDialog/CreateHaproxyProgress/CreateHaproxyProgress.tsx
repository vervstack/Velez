import {useQueryClient} from "@tanstack/react-query"

import type {FinalStepProps} from "@/dialogs/StepsDialog/Step.ts"
import {vcnService} from "@/processes/api/vcn.ts"
import {HAPROXY_INSTANCES_QUERY_KEY, useCreateHaproxyInstanceMutation} from "@/processes/queries/haproxy.ts"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"
import {
    createRequest,
    type HaproxyWizardContext,
    networkMode,
} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"

export default function CreateHaproxyProgress({context, onClose}: FinalStepProps<HaproxyWizardContext>) {
    const queryClient = useQueryClient()
    const createInstance = useCreateHaproxyInstanceMutation()
    const request = createRequest(context)
    const isTailscale = networkMode(context) === "tailscale"

    function start() {
        return createInstance.mutateAsync(request)
    }

    function refresh() {
        void queryClient.invalidateQueries({queryKey: HAPROXY_INSTANCES_QUERY_KEY})
        void queryClient.invalidateQueries({queryKey: ["services"]})
    }

    async function connectSidecar() {
        refresh()
        try {
            await vcnService.connectService({
                serviceName: `haproxy_${request.name}`,
                resolveServiceContainer: true,
                authKey: context.authKey.trim() || undefined,
                loginServerUrl: context.loginServerUrl.trim() || undefined,
                hostname: context.hostname.trim() || undefined,
            })
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error)
            throw new Error(`HAProxy was created, but the Tailscale sidecar failed to connect. ${message}`)
        }
    }

    return (
        <TaskProgressScreen title="Creating HAProxy instance" metaLine={request.name ?? ""}
                            metaLineSecondary={isTailscale ? "Connect Tailscale after provisioning" : undefined}
                            start={start} onTaskDone={isTailscale ? connectSidecar : undefined}
                            onSuccess={refresh} onClose={onClose}/>
    )
}
