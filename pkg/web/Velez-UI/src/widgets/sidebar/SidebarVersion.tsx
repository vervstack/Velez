import cls from '@/widgets/sidebar/SidebarVersion.module.css';
import SkeletonLoader from '@/components/base/SkeletonLoader';
import {useVersionQuery} from '@/processes/queries/version.ts';

export default function SidebarVersion() {
    const query = useVersionQuery();

    if (query.isLoading) {
        return (
            <div className={cls.SidebarVersionContainer}>
                <SkeletonLoader shape="line" width="6rem"/>
            </div>
        );
    }

    const version = query.isError ? undefined : query.data?.version;

    return (
        <div className={cls.SidebarVersionContainer}>
            <span className={cls.VersionText}>Velez · {version || 'unknown'}</span>
        </div>
    );
}
