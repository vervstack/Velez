import {useEffect, useState} from "react";

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {ListServices} from "@/processes/api/control_plane.ts";

import {Service} from "@/model/services/Services";

import ServiceCard from "@/components/service/ServiceCard";
import useSettings from "@/app/settings/state.ts";
import Loader from "@/components/Loader.tsx";

export default function ControlPlanePage() {
    const [activeComponents, setActiveComponents] =
        useState<Service[]>([]);
    const [inactiveComponents, setInactiveComponents] =
        useState<Service[]>([]);


    const [isLoading, setIsLoading] = useState(true)

    const settings = useSettings();

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
                    .map((v, idx) =>
                        <div
                            className={cls.ServiceCardWrapper}
                            key={v.title + idx}
                            onClick={() => {
                                // TODO disabled for now. Use portainer
                                // navigate(Routes.Smerd + '/' + v.title)
                            }}
                        >
                            <ServiceCard
                                disabled={true}
                                {...v}/>
                        </div>)
                }
            </div>
        </div>
    )
}
