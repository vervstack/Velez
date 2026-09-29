import {useEffect} from "react"

import type {RegisterContainerRequest, RegisterContainerResponse} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {queryClient} from "@/app/queryClient.ts"
import {CONTAINER_QUERY_KEY, CONTAINERS_QUERY_KEY, RegisterContainerMutation} from "@/processes/queries/containers.ts"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

const SERVICES_QUERY_KEY = ["services"]

interface Props {
    request: RegisterContainerRequest
    containerName?: string
}

export default function RegisterProgress({request, containerName}: Props) {
    const toaster = useToaster()
    const {LockClosing, UnlockClosing, CloseDialog} = useDialog()
    const registerContainer = RegisterContainerMutation()
    const serviceName = request.serviceName ?? ""

    useEffect(() => {
        LockClosing()
        return UnlockClosing
    }, [])

    function handleStart(): Promise<RegisterContainerResponse> {
        return registerContainer.mutateAsync(request)
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey: CONTAINERS_QUERY_KEY})
        queryClient.invalidateQueries({queryKey: CONTAINER_QUERY_KEY})
        queryClient.invalidateQueries({queryKey: SERVICES_QUERY_KEY})
        toaster.bake({title: "Container registered", description: serviceName, level: "Info"})
    }

    function handleClose() {
        UnlockClosing()
        CloseDialog()
    }

    return (
        <TaskProgressScreen
            title="Registering container"
            metaLine={serviceName}
            metaLineSecondary={containerName}
            start={handleStart}
            onSuccess={handleSuccess}
            onClose={handleClose}
        />
    )
}
