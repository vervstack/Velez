import cls from "@/widgets/settings/SandboxSettings/components/WhitelistedImages/WhitelistedImages.module.css"
import {useListWhitelistedImagesQuery} from "@/processes/queries/settings.ts"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import WhitelistedImageRow from "@/widgets/settings/SandboxSettings/components/WhitelistedImages/components/WhitelistedImageRow/WhitelistedImageRow.tsx"

const DESCRIPTION = "Images allowed to run with elevated access (privileged, devices, capabilities)."

export default function WhitelistedImages() {
    const imagesQuery = useListWhitelistedImagesQuery()

    function handleRetry() {
        imagesQuery.refetch()
    }

    function renderContent() {
        if (imagesQuery.isLoading) {
            return <SkeletonLoader shape="block" height="5rem"/>
        }
        if (imagesQuery.isError) {
            return <QueryErrorState message="Failed to load whitelisted images." onRetry={handleRetry}/>
        }
        return (imagesQuery.data?.images ?? []).map((image) => (
            <WhitelistedImageRow key={image.id || image.repository} image={image}/>
        ))
    }

    return (
        <div className={cls.WhitelistedImagesContainer}>
            <div className={cls.Title}>Whitelisted images</div>
            <p className={cls.Description}>{DESCRIPTION}</p>
            <div className={cls.List}>{renderContent()}</div>
        </div>
    )
}
