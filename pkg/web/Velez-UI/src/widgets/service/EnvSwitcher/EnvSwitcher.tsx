import { useState } from 'react'

import cls from '@/widgets/service/EnvSwitcher/EnvSwitcher.module.css'

import { useListServiceEnvsQuery } from '@/processes/queries/services'
import type { ServiceEnvironment } from '@/model/service_page/ServicePageModel'

interface EnvSwitcherProps {
    serviceId: string
}

export default function EnvSwitcher({ serviceId }: EnvSwitcherProps) {
    const { data: envs = [] } = useListServiceEnvsQuery(serviceId)

    const [activeEnv, setActiveEnv] = useState<string>(() =>
        envs.length > 0 ? envs[0].id : 'prod'
    )

    if (envs.length === 0) {
        return (
            <div className={cls.EnvSwitcherContainer}>
                <p className={cls.Empty}>No environment data available</p>
            </div>
        )
    }

    return (
        <div className={cls.EnvSwitcherContainer}>
            <div className={cls.CardsWrapper}>
                {envs.map(function renderCard(env: ServiceEnvironment) {
                    const isActive = env.id === activeEnv

                    function handleClick() {
                        setActiveEnv(env.id)
                    }

                    return (
                        <div
                            key={env.id}
                            className={`${cls.Card} ${isActive ? cls.active : ''}`}
                            onClick={handleClick}
                        >
                            <div className={cls.CardHeader}>
                                <span className={`${cls.StatusDot} ${cls[`status_${env.status}`]}`} />
                                <span className={cls.Label}>{env.label}</span>
                                {isActive && <span className={cls.ViewingBadge}>viewing</span>}
                            </div>
                            <div className={cls.CardMeta}>
                                <span className={cls.Version}>{env.version}</span>
                                <span className={cls.DeployedAgo}>{env.deployedAgo}</span>
                            </div>
                        </div>
                    )
                })}
            </div>
        </div>
    )
}
