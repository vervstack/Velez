import {ConfirmDialog} from "@vervstack/chures"

import type {DockerContainer} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {queryClient} from "@/app/queryClient.ts"
import {
    CONTAINER_QUERY_KEY,
    CONTAINERS_QUERY_KEY,
    FinishOnboardingMutation,
} from "@/processes/queries/containers.ts"
import Button from "@/components/base/Button.tsx"

interface Props {
    container: DockerContainer
}

export default function FinishOnboardingButton({container}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const finishOnboarding = FinishOnboardingMutation()
    const name = container.name || container.id || "The container"

    function handleConfirm() {
        const environment = useEnvironmentStore.getState().selectedEnvironment

        finishOnboarding.mutateAsync({containerId: container.id ?? "", environment})
            .then(function handleDone() {
                queryClient.invalidateQueries({queryKey: CONTAINERS_QUERY_KEY})
                queryClient.invalidateQueries({queryKey: CONTAINER_QUERY_KEY})
                toaster.bake({title: "Onboarding finished", description: name, level: "Info"})
                CloseDialog()
            })
            .catch(function handleError(err: Error) {
                toaster.catchGrpc(err)
            })
    }

    function handleClick() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Finish onboarding?"
                message={`${name} will be removed. Its volumes are kept, they belong to the new container now.`}
                confirmLabel="Remove"
                onConfirm={handleConfirm}
                onClose={CloseDialog}
            />
        )
    }

    return <Button variant="danger" sm onClick={handleClick}>Finish onboarding</Button>
}
