import cn from "classnames";

import {Tooltip} from "react-tooltip";
import {VervServiceState} from "@vervstack/velez";

import cls from '@/components/service/ServiceCard.module.css'
import RocketSVG from "@/assets/icons/Rocket.svg"

import {Service} from "@/model/services/Services";
import ActivityPoint from "@/components/base/ActivityPoint.tsx";

interface ServiceCardProps extends Service {
    disabled?: boolean

    isDeployRunning?: boolean
    onDeploy?: () => void
}

export default function ServiceCard(serviceProps: ServiceCardProps) {
    return (
        <div className={cn(cls.CardContainer, {
            [cls.disabled]: serviceProps.disabled,
        })}>
            <div className={cls.CardContent}>
                <div className={cn(cls.CardTop, {
                    [cls.disabled]: serviceProps.disabled,
                })}>
                    <div className={cls.ServiceIcon}>
                        <img src={serviceProps.icon} alt={serviceProps.icon}/>
                    </div>

                    <div className={cls.Tittle}>{serviceProps.title}</div>

                    {
                        serviceProps.webLink && (<div
                            className={cls.ActionButton}
                            data-tooltip-id={"ServiceCardTooltip"}
                            data-tooltip-content="Open in new window"
                            data-tooltip-place="left"
                            onClick={() => window.open(serviceProps.webLink, '_blank')}
                        >
                        <span
                            className="material-symbols-outlined"
                            children={"open_in_new"}/>
                        </div>)
                    }

                    {
                        serviceProps.disabled && serviceProps.onDeploy && (
                            <div
                                className={cn(cls.ActionButton, {
                                    [cls.disabled]: serviceProps.isDeployRunning,
                                })}
                                data-tooltip-id={"ServiceCardTooltip"}
                                data-tooltip-content="Deploy on this node"
                                data-tooltip-place="left"
                                onClick={serviceProps.onDeploy}
                            >
                                <div className={cls.CardButton}>
                                    <img src={RocketSVG} alt={'/'}/>
                                </div>
                            </div>)
                    }

                </div>
                <div className={cls.CardBottom}>
                    <div className={cls.Content}>
                        {serviceProps.description}
                    </div>
                </div>
            </div>

            <div className={cls.ActivityPoint}>
                <ActivityPoint
                    isInactive={serviceProps.disabled || serviceProps.state != VervServiceState.running}/>
            </div>

            <Tooltip
                id={"ServiceCardTooltip"}
            />
        </div>
    )
}

