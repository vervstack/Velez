import cls from '@/components/node/SkeletonNodeRow.module.css';
import SkeletonLoader from '@/components/base/SkeletonLoader';

export default function SkeletonNodeRow() {
    return (
        <div className={cls.SkeletonNodeRowContainer}>
            <SkeletonLoader shape="block" width="0.625rem" height="0.625rem" />
            <div className={cls.info}>
                <SkeletonLoader shape="line" width="6rem" height="0.5rem" />
                <SkeletonLoader shape="line" width="4rem" height="0.4375rem" />
            </div>
        </div>
    );
}
