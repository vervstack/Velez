import cls from '@/components/node/SkeletonNodeCard.module.css';
import SkeletonLoader from '@/components/base/SkeletonLoader';

export default function SkeletonNodeCard() {
    return (
        <div className={cls.SkeletonNodeCardContainer}>
            <div className={cls.identity}>
                <SkeletonLoader shape="line" width="6rem" height="0.625rem" />
                <SkeletonLoader shape="line" width="9rem" height="0.5rem" />
                <SkeletonLoader shape="line" width="5rem" height="0.5rem" />
            </div>

            <div className={cls.metric}>
                <SkeletonLoader shape="line" width="100%" height="2.5rem" />
            </div>

            <div className={cls.metric}>
                <SkeletonLoader shape="line" width="100%" height="2.5rem" />
            </div>

            <div className={cls.servicesBlock}>
                <SkeletonLoader shape="block" width="2rem" height="1.375rem" />
                <SkeletonLoader shape="line" width="3rem" height="0.5rem" />
            </div>

            <div className={cls.actions}>
                <SkeletonLoader shape="block" width="2rem" height="2rem" />
                <SkeletonLoader shape="block" width="2rem" height="2rem" />
            </div>
        </div>
    );
}