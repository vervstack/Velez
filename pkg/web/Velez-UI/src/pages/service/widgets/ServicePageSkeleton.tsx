import cls from "@/pages/service/widgets/ServicePageSkeleton.module.css";
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx";

export default function ServicePageSkeleton() {
    return (
        <div className={cls.ServicePageSkeletonContainer}>
            <div className={cls.ContentWrapper}>
                <SkeletonLoader shape="block" width="100%" height="6rem"/>
                <SkeletonLoader shape="block" width="100%" height="8rem"/>
                <SkeletonLoader shape="block" width="100%" height="10rem"/>
            </div>
        </div>
    );
}
