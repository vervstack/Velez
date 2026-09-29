import cls from "@/dialogs/CreateServiceDialog/screens/PostgresAdoptScreen/PostgresAdoptScreen.module.css"
import type {DockerContainer} from "@/app/api/velez"
import {useGetContainerQuery} from "@/processes/queries/containers.ts"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import AdoptForm from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx"
import {isPgLoginMissing} from "@/dialogs/CreateServiceDialog/processes/pgLogin.ts"

interface Props {
    container: DockerContainer
    onBusyChange(isBusy: boolean): void
}

export default function PostgresAdoptScreen({container, onBusyChange}: Props) {
    const containerId = container.id ?? ""
    const {data, isLoading, isError, refetch} = useGetContainerQuery(containerId, containerId !== "")

    if (isLoading) {
        return (
            <div className={cls.PostgresAdoptScreenContainer}>
                <SkeletonLoader shape="block" height="2.5rem"/>
                <SkeletonLoader shape="block" height="2.5rem"/>
                <SkeletonLoader shape="block" height="5rem"/>
            </div>
        )
    }

    if (isError || !data) {
        return (
            <div className={cls.PostgresAdoptScreenContainer}>
                <QueryErrorState message="Failed to load the container." onRetry={refetch}/>
            </div>
        )
    }

    return (
        <AdoptForm
            container={container}
            pattern="postgres"
            isPgLoginRequired={isPgLoginMissing(data.env)}
            onBusyChange={onBusyChange}
        />
    )
}
