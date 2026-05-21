import { useState } from 'react';

import cls from '@/widgets/service/EnvSwitcher/EnvSwitcher.module.css';
import EnvCard from '@/components/complex/EnvCard/EnvCard.tsx';

import { useListServiceEnvsQuery } from '@/processes/queries/services';
import type { ServiceEnvironment } from '@/model/service_page/ServicePageModel';

interface EnvSwitcherProps {
    serviceName: string;
}

export default function EnvSwitcher({ serviceName }: EnvSwitcherProps) {
    const { data: envs = [] } = useListServiceEnvsQuery(serviceName);

    const [activeEnv, setActiveEnv] = useState<string>(() =>
        envs.length > 0 ? envs[0].id : 'prod'
    );

    if (envs.length === 0) {
        return (
            <div className={cls.EnvSwitcherContainer}>
                <p className={cls.Empty}>No environment data available</p>
            </div>
        );
    }

    return (
        <div className={cls.EnvSwitcherContainer}>
            <div className={cls.CardsWrapper}>
                {envs.map(function renderCard(env: ServiceEnvironment) {
                    function handleClick() {
                        setActiveEnv(env.id);
                    }

                    return (
                        <EnvCard
                            key={env.id}
                            env={env}
                            isActive={env.id === activeEnv}
                            onClick={handleClick}
                        />
                    );
                })}
            </div>
        </div>
    );
}
