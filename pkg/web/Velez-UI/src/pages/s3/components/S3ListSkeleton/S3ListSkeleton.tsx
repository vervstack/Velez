import cls from "@/pages/s3/components/S3ListSkeleton/S3ListSkeleton.module.css"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"

const ROW_KEYS = ["first", "second", "third"]

function renderRow(key: string) {
    return (
        <div key={key} className={cls.Row}>
            <SkeletonLoader shape="line" height="1rem"/>
        </div>
    )
}

export default function S3ListSkeleton() {
    return (
        <div className={cls.S3ListSkeletonContainer}>
            {ROW_KEYS.map(renderRow)}
        </div>
    )
}
