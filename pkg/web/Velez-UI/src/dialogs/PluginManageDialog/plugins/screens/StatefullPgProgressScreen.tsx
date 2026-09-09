import {EnablePluginResponse, EnableStatefullCluster} from "@/app/api/velez"
import {queryClient} from "@/app/queryClient.ts"
import {StatefullPgContext} from "@/dialogs/PluginManageDialog/plugins/StatefullPgContext.ts"
import TaskProgressScreen from "@/dialogs/PluginManageDialog/plugins/screens/TaskProgressScreen.tsx"
import {controlPlaneService} from "@/processes/api/control_plane.ts"
import {NodeHardwareQuery} from "@/processes/queries/control_plane.ts"

interface Props {
    context: StatefullPgContext

    onClose(): void
}

export default function StatefullPgProgressScreen({context, onClose}: Props) {
    const hardwareQuery = NodeHardwareQuery()
    const nodeRegion = hardwareQuery.data?.nodeRegion

    function handleStart(): Promise<EnablePluginResponse> {
        const payload: EnableStatefullCluster = {
            isExposePort: context.exposePort,
            exposeToPort: context.exposePort ? context.portNumber : undefined,
        }
        return controlPlaneService.enableStatefullPgCluster(payload)
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey: ["plugins"]})
        queryClient.invalidateQueries({queryKey: ["nodes"]})
    }

    return (
        <TaskProgressScreen
            title="Enabling cluster mode"
            metaLine="PostgreSQL cluster"
            metaLineSecondary={nodeRegion}
            start={handleStart}
            onSuccess={handleSuccess}
            onClose={onClose}
        />
    )
}
