import { useState } from 'react';

import cls from '@/widgets/service/EnvSwitcher/EnvSwitcher.module.css';
import EnvCard from '@/components/complex/EnvCard/EnvCard.tsx';

import { ListEnvironmentsQuery } from '@/processes/queries/control_plane';
import type { ServiceEnvironment } from '@/model/service_page/ServicePageModel';

export default function EnvSwitcher() {
    const { data } = ListEnvironmentsQuery();
    const envs: ServiceEnvironment[] = (data?.environments ?? []).map(function toEnv(name) {
        return { id: name, label: name, status: 'running', version: '', deployedAgo: '', health: 'healthy' };
    });

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
