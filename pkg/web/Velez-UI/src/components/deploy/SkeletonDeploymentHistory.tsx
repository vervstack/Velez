import cls from '@/components/deploy/SkeletonDeploymentHistory.module.css';
import SkeletonLoader from '@/components/base/SkeletonLoader';

export default function SkeletonDeploymentHistory() {
    return (
        <div className={cls.SkeletonDeploymentHistoryContainer}>
            <div className={cls.item}>
                <SkeletonLoader shape="line" width="25%" height="0.8rem" />
                <SkeletonLoader shape="line" width="50%" height="0.7rem" />
            </div>
            <div className={cls.item}>
                <SkeletonLoader shape="line" width="25%" height="0.8rem" />
                <SkeletonLoader shape="line" width="50%" height="0.7rem" />
            </div>
            <div className={cls.item}>
                <SkeletonLoader shape="line" width="25%" height="0.8rem" />
                <SkeletonLoader shape="line" width="50%" height="0.7rem" />
            </div>
        </div>
    );
}
