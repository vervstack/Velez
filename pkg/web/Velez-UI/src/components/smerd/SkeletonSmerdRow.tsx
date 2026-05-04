import cls from '@/components/smerd/SkeletonSmerdRow.module.css';
import SkeletonLoader from '@/components/base/SkeletonLoader';

export default function SkeletonSmerdRow() {
    return (
        <div className={cls.SkeletonSmerdRowContainer}>
            <SkeletonLoader shape="block" width="0.5rem" height="0.5rem" />
            <SkeletonLoader shape="line" width="20%" height="0.9rem" />
            <SkeletonLoader shape="line" width="40%" height="0.8rem" />
            <SkeletonLoader shape="line" width="15%" height="0.72rem" />
        </div>
    );
}
