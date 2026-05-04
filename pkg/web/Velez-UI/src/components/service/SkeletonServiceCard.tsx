import cls from '@/components/service/SkeletonServiceCard.module.css';
import SkeletonLoader from '@/components/base/SkeletonLoader';

export default function SkeletonServiceCard() {
    return (
        <div className={cls.SkeletonServiceCardContainer}>
            <SkeletonLoader shape="line" width="80%" height="0.95rem" />
            <SkeletonLoader shape="line" width="60%" height="0.72rem" />
            <SkeletonLoader shape="line" width="50%" height="0.7rem" />
        </div>
    );
}
