import cls from "@/dialogs/CreateServiceDialog/screens/RunnerAdoptScreen/RunnerAdoptScreen.module.css"
import type {DockerContainer, RunnerProvider} from "@/app/api/velez"
import {useGetContainerQuery} from "@/processes/queries/containers.ts"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import AdoptForm from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx"
import {initialRunnerForm} from "@/dialogs/CreateServiceDialog/processes/runnerDefaults.ts"
import type {OnboardFormState} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"

interface Props {
    container: DockerContainer
    initialProvider?: RunnerProvider
    onFormChange(form: OnboardFormState): void
}

export default function RunnerAdoptScreen({container, initialProvider, onFormChange}: Props) {
    const containerId = container.id ?? ""
    const {data, isLoading, isError, refetch} = useGetContainerQuery(containerId, containerId !== "")

    if (isLoading) {
        return (
            <div className={cls.RunnerAdoptScreenContainer}>
                <SkeletonLoader shape="block" height="2.5rem"/>
                <SkeletonLoader shape="block" height="2.5rem"/>
                <SkeletonLoader shape="block" height="5rem"/>
            </div>
        )
    }

    if (isError || !data) {
        return (
            <div className={cls.RunnerAdoptScreenContainer}>
                <QueryErrorState message="Failed to load the container." onRetry={refetch}/>
            </div>
        )
    }

    const prefilled = initialRunnerForm(data)

    return (
        <AdoptForm
            container={container}
            pattern="runner"
            initialRunner={{...prefilled, provider: prefilled.provider ?? initialProvider}}
            isRegistrationTokenFound={Boolean(data.suggestedRunnerDefaults?.isRegistrationTokenFound)}
            onFormChange={onFormChange}
        />
    )
}
