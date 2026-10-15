import cn from 'classnames';

import cls from '@/components/complex/EnvCard/EnvCard.module.css';
import type { ServiceEnvironment } from '@/model/service_page/ServicePageModel';
import { shortContainerId } from '@/processes/mappings/containers';

interface EnvCardProps {
    env: ServiceEnvironment;
    isActive?: boolean;
    isLinked?: boolean;
    onClick?: () => void;
}

export default function EnvCard({ env, isActive = false, isLinked = false, onClick }: EnvCardProps) {
    const containerId = shortContainerId(env.containerId);

    return (
        <div
            className={cn(cls.EnvCardContainer, isActive && cls.active, isLinked && cls.linked)}
            onClick={onClick}
        >
            <div className={cls.CardHeaderWrapper}>
                <span className={cn(cls.StatusDot, cls['status_' + env.status])} />
                {env.label && <span className={cls.Label}>{env.label}</span>}
                <span className={cls.DeployedAgo}>{env.deployedAgo}</span>
            </div>
            <dl className={cls.InfoWrapper}>
                {containerId && <><dt className={cls.InfoKey}>id</dt><dd className={cls.InfoValue}>{containerId}</dd></>}
                {env.containerName && <><dt className={cls.InfoKey}>name</dt><dd className={cls.InfoValue}>{env.containerName}</dd></>}
                {env.version && <><dt className={cls.InfoKey}>image</dt><dd className={cls.InfoValue}>{env.version}</dd></>}
            </dl>
        </div>
    );
}
