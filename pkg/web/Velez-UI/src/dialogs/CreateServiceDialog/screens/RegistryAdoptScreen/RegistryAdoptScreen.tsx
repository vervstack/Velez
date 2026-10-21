import cls from "@/dialogs/CreateServiceDialog/screens/RegistryAdoptScreen/RegistryAdoptScreen.module.css"
import type {DockerContainer} from "@/app/api/velez"
import {useGetContainerQuery} from "@/processes/queries/containers.ts"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import AdoptForm from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx"
import {isRegistryLoginRequired} from "@/dialogs/CreateServiceDialog/processes/registryLogin.ts"
import type {OnboardFormState} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"

interface Props {
    container: DockerContainer
    onFormChange(form: OnboardFormState): void
}

export default function RegistryAdoptScreen({container, onFormChange}: Props) {
    const containerId = container.id ?? ""
    const {data, isLoading, isError, refetch} = useGetContainerQuery(containerId, containerId !== "")

    if (isLoading) {
        return (
            <div className={cls.RegistryAdoptScreenContainer}>
                <SkeletonLoader shape="block" height="2.5rem"/>
                <SkeletonLoader shape="block" height="2.5rem"/>
                <SkeletonLoader shape="block" height="5rem"/>
            </div>
        )
    }

    if (isError || !data) {
        return (
            <div className={cls.RegistryAdoptScreenContainer}>
                <QueryErrorState message="Failed to load the container." onRetry={refetch}/>
            </div>
        )
    }

    return (
        <AdoptForm
            container={container}
            pattern="registry"
            isRegistryLoginRequired={isRegistryLoginRequired(data.env)}
            onFormChange={onFormChange}
        />
    )
}
