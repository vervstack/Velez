import cls from "@/dialogs/CreateServiceDialog/components/ContainerPreviewView/ContainerPreviewView.module.css"
import type {ContainerPreview} from "@/app/api/velez"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import PreviewSectionView
    from "@/dialogs/CreateServiceDialog/components/PreviewSectionView/PreviewSectionView.tsx"
import {PreviewSection, previewSectionsOf} from "@/dialogs/CreateServiceDialog/processes/registerPreview.ts"

export interface ContainerPreviewState {
    preview?: ContainerPreview
    isLoading: boolean
    isError: boolean
    isDisabled: boolean
    onRetry(): void
}

export default function ContainerPreviewView(
    {preview, isLoading, isError, isDisabled, onRetry}: ContainerPreviewState,
) {
    if (isDisabled) return null

    if (isLoading) {
        return (
            <div className={cls.ContainerPreviewViewContainer}>
                <SkeletonLoader shape="block" height="1.5rem"/>
                <SkeletonLoader shape="block" height="1.5rem"/>
                <SkeletonLoader shape="block" height="1.5rem"/>
            </div>
        )
    }

    if (isError) {
        return (
            <div className={cls.ContainerPreviewViewContainer}>
                <QueryErrorState message="Failed to load the preview." onRetry={onRetry}/>
            </div>
        )
    }

    const sections = preview ? previewSectionsOf(preview) : []

    if (sections.length === 0) {
        return (
            <div className={cls.ContainerPreviewViewContainer}>
                <span className={cls.Empty}>Nothing to show for this container.</span>
            </div>
        )
    }

    return (
        <div className={cls.ContainerPreviewViewContainer}>
            {sections.map(function renderSection(section: PreviewSection) {
                return <PreviewSectionView key={section.key} section={section}/>
            })}
        </div>
    )
}
