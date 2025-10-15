import {useEffect, useState} from "react";

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {ListServices} from "@/processes/api/control_plane.ts";

import {Service} from "@/model/services/Services";

import ServiceCard from "@/components/service/ServiceCard";
import useSettings from "@/app/settings/state.ts";
import Loader from "@/components/Loader.tsx";
import {useNavigate} from "react-router-dom";
import {Routes} from "@/app/router/Router.tsx";
import {fromProto} from "@/model/smerds/Smerds.ts";

export default function ControlPlanePage() {
    const [activeComponents, setActiveComponents] =
        useState<Service[]>([]);
    const [inactiveComponents, setInactiveComponents] =
        useState<Service[]>([]);


    const [isLoading, setIsLoading] = useState(true)

    const settings = useSettings();
    const navigate = useNavigate();

    useEffect(() => {
        setIsLoading(true)
        ListServices(settings.initReq())
            .then((r) => {
                setActiveComponents(r.active)
                setInactiveComponents(r.inactive)
            })
            .then(() => setIsLoading(false))
    }, []);


    if (isLoading) {
        return (
            <div className={cls.ControlPlaneContainer}>
                <Loader/>
            </div>
        )
    }

    return (
        <div className={cls.ControlPlaneContainer}>
            <div className={cls.ServicesBlock}>
                {activeComponents
                    .map((v, idx) =>
                        <div
                            className={cls.ServiceCardWrapper}
                            key={v.title + idx}
                            onClick={() => {
                                // TODO disabled for now. Use portainer
                                // navigate(Routes.Smerd + '/' + v.title)
                            }}

                        >
                            <ServiceCard disabled={false} {...v}/>
                        </div>)
                }
            </div>

            <div className={cls.ServicesBlock}>
                {inactiveComponents
                    .map((v: Service, idx) =>
                        <div
                            className={cls.ServiceCardWrapper}
                            key={v.title + idx}
                        >
                            <ServiceCard
                                onClickConstructor={v.smerdConstructor !== undefined ? () => {
                                    navigate(Routes.Deploy,
                                        {state: {data: fromProto(v.smerdConstructor)}})
                                } : undefined}
                                disabled={true}
                                {...v}/>
                        </div>)
                }
            </div>
        </div>
    )
}
