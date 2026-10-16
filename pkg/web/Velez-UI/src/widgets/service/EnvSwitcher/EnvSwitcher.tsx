import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import cls from '@/widgets/service/EnvSwitcher/EnvSwitcher.module.css';
import EnvCard from '@/components/complex/EnvCard/EnvCard.tsx';

import { Routes } from '@/app/router/Routes';
import { useListServiceEnvsQuery } from '@/processes/queries/services.ts';
import { sidecarToEnvironment } from '@/processes/mappings/serviceSidecars';
import type { ServiceEnvironment, ServiceSidecarView } from '@/model/service_page/ServicePageModel';

interface EnvSwitcherProps {
    serviceName?: string;
    sidecars?: ServiceSidecarView[];
}

export default function EnvSwitcher({ serviceName = '', sidecars }: EnvSwitcherProps) {
    const navigate = useNavigate();
    const { data } = useListServiceEnvsQuery(serviceName);
    const envs: ServiceEnvironment[] = data ?? [];
    const sidecarEnvs = useMemo(() => (sidecars ?? []).map(sidecarToEnvironment), [sidecars]);

    const [activeEnv, setActiveEnv] = useState<string>(() =>
        envs.length > 0 ? envs[0].id : 'prod'
    );

    if (envs.length === 0 && sidecarEnvs.length === 0) {
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
                        if (env.containerId) {
                            navigate(Routes.Container + '/' + env.containerId);
                        }
                    }

                    return (
                        <EnvCard
                            key={env.id}
                            env={env}
                            isActive={env.id === activeEnv}
                            isLinked={!!env.containerId}
                            onClick={handleClick}
                        />
                    );
                })}
                {sidecarEnvs.map(function renderSidecarCard(sidecarEnv: ServiceEnvironment) {
                    function handleSidecarClick() {
                        navigate(Routes.Container + '/' + sidecarEnv.containerId);
                    }

                    return (
                        <EnvCard
                            key={'sidecar-' + sidecarEnv.containerId}
                            env={sidecarEnv}
                            isActive={false}
                            isLinked
                            onClick={handleSidecarClick}
                        />
                    );
                })}
            </div>
        </div>
    );
}
