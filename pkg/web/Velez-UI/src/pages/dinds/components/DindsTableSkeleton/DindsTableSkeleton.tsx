import cls from "@/pages/dinds/components/DindsTableSkeleton/DindsTableSkeleton.module.css"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"

const ROW_KEYS = ["first", "second", "third"]

function renderRow(key: string) {
    return (
        <div key={key} className={cls.Row}>
            <SkeletonLoader shape="line" height="1rem"/>
        </div>
    )
}

export default function DindsTableSkeleton() {
    return (
        <div className={cls.DindsTableSkeletonContainer}>
            {ROW_KEYS.map(renderRow)}
        </div>
    )
}
