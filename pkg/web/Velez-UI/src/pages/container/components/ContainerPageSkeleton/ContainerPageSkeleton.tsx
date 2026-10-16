import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import cls from "@/pages/container/components/ContainerPageSkeleton/ContainerPageSkeleton.module.css"

export default function ContainerPageSkeleton() {
    return (
        <div className={cls.ContainerPageSkeletonContainer}>
            <div className={cls.Header}>
                <SkeletonLoader shape="line" width="12rem" height="2rem"/>
                <SkeletonLoader shape="block" width="4rem" height="1.2rem"/>
            </div>
            <div className={cls.Box}>
                <SkeletonLoader shape="line" width="6rem" height="0.8rem"/>
                <div className={cls.Body}>
                    <SkeletonLoader shape="line" width="60%" height="0.9rem"/>
                    <SkeletonLoader shape="line" width="40%" height="0.9rem"/>
                    <SkeletonLoader shape="line" width="50%" height="0.9rem"/>
                </div>
            </div>
        </div>
    )
}
