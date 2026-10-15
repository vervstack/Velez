import cn from 'classnames';

import cls from '@/components/complex/EnvCard/EnvCard.module.css';
import type { ServiceEnvironment } from '@/model/service_page/ServicePageModel';

interface EnvCardProps {
    env: ServiceEnvironment;
    isActive?: boolean;
    isLinked?: boolean;
    onClick?: () => void;
}

export default function EnvCard({ env, isActive = false, isLinked = false, onClick }: EnvCardProps) {
    return (
        <div
            className={cn(cls.EnvCardContainer, isActive && cls.active, isLinked && cls.linked)}
            onClick={onClick}
        >
            <div className={cls.CardHeaderWrapper}>
                <span className={cn(cls.StatusDot, cls['status_' + env.status])} />
                {env.label && <span className={cls.Label}>{env.label}</span>}
                <span className={cls.Version}>{env.version}</span>
            </div>
            <div className={cls.CardMetaWrapper}>
                <span className={cls.DeployedAgo}>{env.deployedAgo}</span>
            </div>
        </div>
    );
}
