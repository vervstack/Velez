import cls from "@/pages/transfer/components/TransfersListSkeleton/TransfersListSkeleton.module.css"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"

const ROW_KEYS = ["first", "second", "third"]

function renderRow(key: string) {
    return (
        <div key={key} className={cls.Row}>
            <SkeletonLoader shape="line" height="1rem"/>
        </div>
    )
}

export default function TransfersListSkeleton() {
    return (
        <div className={cls.TransfersListSkeletonContainer}>
            {ROW_KEYS.map(renderRow)}
        </div>
    )
}
