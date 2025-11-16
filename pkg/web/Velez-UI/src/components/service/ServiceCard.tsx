import cn from "classnames";

import cls from '@/components/service/ServiceCard.module.css'
import RocketSVG from "@/assets/icons/Rocket.svg"

import {Service} from "@/model/services/Services";
import ActivityPoint from "@/components/base/ActivityPoint.tsx";
import Toggle from "@/components/base/Toggle.tsx";
import {useState} from "react";

interface ServiceCardProps extends Service {
    disabled?: boolean

    onClickConstructor?: () => void
    isTogglable?: boolean
}

export default function ServiceCard({
                                        title,
                                        icon,
                                        webLink,
                                        description,
                                        disabled,

                                        onClickConstructor,

                                        isTogglable
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
                        webLink && (<div
                            className={cls.ExternalLink}
                            data-tooltip-id={"tooltip"}
                            data-tooltip-content="Open in new window"
                            data-tooltip-place="left"
                            onClick={() => window.open(webLink, '_blank')}
                        >
                        <span
                            className="material-symbols-outlined"
                            children={"open_in_new"}/>
                        </div>)
                    }

                    {
                        onClickConstructor && (<div
                            className={cls.ExternalLink}
                            data-tooltip-id={"tooltip"}
                            data-tooltip-content="Deploy on this node"
                            data-tooltip-place="left"
                            onClick={onClickConstructor}
                        >
                            <div className={cls.CardButton}>
                                <img src={RocketSVG} alt={'/'}/>
                            </div>
                        </div>)
                    }

                    {
                        onClickConstructor && (<div
                            className={cls.ExternalLink}
                            data-tooltip-id={"tooltip"}
                            data-tooltip-content="Deploy on this node"
                            data-tooltip-place="left"
                            onClick={onClickConstructor}
                        >
                            <div className={cls.CardButton}>
                                <img src={RocketSVG} alt={'/'}/>
                            </div>
                        </div>)
                    }

                    {isTogglable && (<ToggleService isToggled={!disabled}/>)}

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
        </div>
    )
}


type ToggleProps = {
    isToggled: boolean
}

function ToggleService({isToggled}: ToggleProps) {
    const [toggled, setToggled] = useState(isToggled)

    function onToggle(newVal: boolean) {
        setToggled(newVal)
    }

    return (<div
        className={cls.ExternalLink}
        data-tooltip-id={"tooltip"}
        data-tooltip-content="Enable"
        data-tooltip-place="left"
    >
        <Toggle
            value={toggled}
            onChange={onToggle}
        />
    </div>)
}
