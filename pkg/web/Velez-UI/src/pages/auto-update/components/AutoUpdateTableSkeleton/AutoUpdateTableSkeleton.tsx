import cls from "@/pages/auto-update/components/AutoUpdateTableSkeleton/AutoUpdateTableSkeleton.module.css"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"

const ROW_KEYS = ["first", "second", "third"]

function renderRow(key: string) {
    return (
        <div key={key} className={cls.Row}>
            <SkeletonLoader shape="line" height="1rem" width="60%"/>
            <SkeletonLoader shape="line" height="1rem" width="80%"/>
            <SkeletonLoader shape="line" height="1.5rem"/>
            <SkeletonLoader shape="line" height="1.5rem" width="3rem"/>
        </div>
    )
}

export default function AutoUpdateTableSkeleton() {
    return (
        <div className={cls.AutoUpdateTableSkeletonContainer}>
            {ROW_KEYS.map(renderRow)}
        </div>
    )
}
