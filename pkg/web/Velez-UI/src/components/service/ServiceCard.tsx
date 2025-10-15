import cn from "classnames";
import {Tooltip} from "react-tooltip";

import cls from '@/components/service/ServiceCard.module.css'
import RocketSVG from "@/assets/icons/Rocket.svg"

import {Service} from "@/model/services/Services";
import ActivityPoint from "@/components/base/ActivityPoint.tsx";

interface ServiceCardProps extends Service {
    disabled?: boolean

    onClickConstructor?: () => void
}

export default function ServiceCard({
                                        title,
                                        icon,
                                        webLink,
                                        description,
                                        disabled,
                                        onClickConstructor
                                    }: ServiceCardProps) {
    return (
        <div className={cn(cls.CardContainer, {
            [cls.disabled]: disabled,
        })}>
            <div className={cls.CardContent}>
                <div className={cls.CardTop}>
                    <div className={cls.ServiceIcon}>
                        <img src={icon} alt={icon}/>
                    </div>

                    <div className={cls.Tittle}>{title}</div>

                    {
                        webLink ? <div
                            className={cls.ExternalLink}
                            data-tooltip-id={"open-external-service-link-" + title}
                            data-tooltip-content="Open in new window"
                            data-tooltip-place="left"
                            onClick={() => window.open(webLink, '_blank')}
                        >
                        <span
                            className="material-symbols-outlined"
                            children={"open_in_new"}/>
                        </div> : null
                    }

                    {
                        onClickConstructor ? <div
                            className={cls.ExternalLink}
                            data-tooltip-id={"open-external-service-link-" + title}
                            data-tooltip-content="Deploy on this node"
                            data-tooltip-place="left"
                            onClick={onClickConstructor}
                        >
                            <div className={cls.CardButton}>
                                <img src={RocketSVG} alt={'/'}/>
                            </div>
                        </div> : null
                    }
                </div>
                <div className={cls.CardBottom}>
                    <div className={cls.Content}>
                        {description}
                    </div>
                </div>
            </div>

            <div className={cls.ActivityPoint}>
                <ActivityPoint isInactive={disabled}/>
            </div>

            <Tooltip
                id={"open-external-service-link-" + title}
            />
        </div>
    )
}
