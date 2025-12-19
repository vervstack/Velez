import cn from "classnames";

import cls from '@/components/service/ServiceCard.module.css'
import RocketSVG from "@/assets/icons/Rocket.svg"

import {Service} from "@/model/services/Services";
import ActivityPoint from "@/components/base/ActivityPoint.tsx";
import {EnableService} from "@/processes/api/control_plane.ts";
import {useCredentialsStore} from "@/app/settings/creds.ts";
import {useState} from "react";

interface ServiceCardProps extends Service {
    disabled?: boolean
    doRefresh?: () => void
}

export default function ServiceCard(serviceProps: ServiceCardProps) {
    const credentialsStore = useCredentialsStore();

    const [isDeployRunning, setIsDeployRunning] = useState(false)

    function enableService() {
        setIsDeployRunning(true)
        EnableService(serviceProps.type, credentialsStore.getInitReq())
            .then(() => {
                if (serviceProps.doRefresh) serviceProps.doRefresh()
            })
            .finally(() => {
                setIsDeployRunning(false)
            })
    }

    return (
        <div className={cn(cls.CardContainer, {
            [cls.disabled]: serviceProps.disabled,
        })}>
            <div className={cls.CardContent}>
                <div className={cls.CardTop}>
                    <div className={cls.ServiceIcon}>
                        <img src={serviceProps.icon} alt={serviceProps.icon}/>
                    </div>

                    <div className={cls.Tittle}>{serviceProps.title}</div>

                    {
                        serviceProps.webLink && (<div
                            className={cls.ExternalLink}
                            data-tooltip-id={"tooltip"}
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
                        serviceProps.disabled && (
                            <div
                                className={cn(cls.ExternalLink, {
                                    [cls.disabled]: isDeployRunning,
                                })}
                                data-tooltip-id={"tooltip"}
                                data-tooltip-content="Deploy on this node"
                                data-tooltip-place="left"
                                onClick={enableService}
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
                    isInactive={serviceProps.disabled}/>
            </div>
        </div>
    )
}

