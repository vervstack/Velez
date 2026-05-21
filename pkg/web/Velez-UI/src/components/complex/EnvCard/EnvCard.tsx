import cls from '@/components/complex/EnvCard/EnvCard.module.css';
import type { ServiceEnvironment } from '@/model/service_page/ServicePageModel';

interface EnvCardProps {
    env: ServiceEnvironment;
    isActive?: boolean;
    onClick?: () => void;
}

export default function EnvCard({ env, isActive = false, onClick }: EnvCardProps) {
    return (
        <div
            className={`${cls.EnvCardContainer} ${isActive ? cls.active : ''}`}
            onClick={onClick}
        >
            <div className={cls.CardHeaderWrapper}>
                <span className={`${cls.StatusDot} ${cls[`status_${env.status}`]}`} />
                <span className={cls.Label}>{env.label}</span>
                {isActive && <span className={cls.ViewingBadge}>viewing</span>}
            </div>
            <div className={cls.CardMetaWrapper}>
                <span className={cls.Version}>{env.version}</span>
                <span className={cls.DeployedAgo}>{env.deployedAgo}</span>
            </div>
        </div>
    );
}
