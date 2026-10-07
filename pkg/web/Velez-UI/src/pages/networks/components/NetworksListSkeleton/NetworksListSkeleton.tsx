import cls from "@/pages/networks/components/NetworksListSkeleton/NetworksListSkeleton.module.css"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"

const ROW_KEYS = ["first", "second", "third"]

function renderRow(key: string) {
    return (
        <div key={key} className={cls.Row}>
            <SkeletonLoader shape="line" height="1rem"/>
        </div>
    )
}

export default function NetworksListSkeleton() {
    return (
        <div className={cls.NetworksListSkeletonContainer}>
            {ROW_KEYS.map(renderRow)}
        </div>
    )
}
