import {useEffect, useState} from "react";
import {useNavigate} from "react-router-dom";

import {CreateSmerdRequest} from "@vervstack/velez"

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {ListServices} from "@/processes/api/control_plane.ts";
import {Service} from "@/model/services/Services";
import ServiceCard from "@/components/service/ServiceCard";
import Loader from "@/components/Loader.tsx";
import {Routes} from "@/app/router/Router.tsx";

import {fromProto} from "@/model/smerds/Smerds.ts";
import {useCredentialsStore} from "@/app/settings/creds.ts";

export default function ControlPlanePage() {
    const [activeComponents, setActiveComponents] =
        useState<Service[]>([]);

    const [inactiveComponents, setInactiveComponents] =
        useState<Service[]>([]);

    useEffect(() => {
        console.log(123)
    }, []);

    const [isLoading, setIsLoading] = useState(true)

    const credentialsStore = useCredentialsStore();
    const navigate = useNavigate();

    useEffect(() => {
        setIsLoading(true)
        ListServices(credentialsStore.getInitReq())
            .then((r) => {
                setActiveComponents(r.active)
                setInactiveComponents(r.inactive)
                console.log(421)
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

    function extractConstructor(constr: CreateSmerdRequest | undefined) {
        if (constr == undefined) return

        return () => {
            navigate(Routes.Deploy, {state: {data: fromProto(constr)}})
        }
    }

    function extractTogglable(smerd: Service): boolean {
        return smerd.togglable
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
                                onClickConstructor={extractConstructor(v.smerdConstructor)}
                                isTogglable={extractTogglable(v)}
                                disabled={true}
                                {...v}/>
                        </div>)
                }
            </div>
        </div>
    )
}
