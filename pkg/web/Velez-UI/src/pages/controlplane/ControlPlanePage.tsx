import {useEffect, useState} from "react";

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {ListServices} from "@/processes/api/control_plane.ts";
import {Service} from "@/model/services/Services";
import ServiceCard from "@/components/service/ServiceCard";
import Loader from "@/components/Loader.tsx";

import {useCredentialsStore} from "@/app/settings/creds.ts";

export default function ControlPlanePage() {
    const [activeComponents, setActiveComponents] =
        useState<Service[]>([]);

    const [inactiveComponents, setInactiveComponents] =
        useState<Service[]>([]);

    const [isLoading, setIsLoading] = useState(true)

    const credentialsStore = useCredentialsStore();

    useEffect(() => {
        setIsLoading(true)
        ListServices(credentialsStore.getInitReq())
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

    // TODO save for future use
    // function extractConstructor(constr: CreateSmerdRequest | undefined) {
    //     if (constr == undefined) return
    //
    //     return () => {
    //         navigate(Routes.Deploy, {state: {data: fromProto(constr)}})
    //     }
    // }

    return (
        <div className={cls.ControlPlaneContainer}>
            <div className={cls.ServicesBlock}>
                {activeComponents
                    .map((v, idx) =>
                        <div
                            className={cls.ServiceCardWrapper}
                            key={v.title + idx}
                            onClick={() => {
                                // TODO when clicked will navigate to service's page
                                // navigate(Routes.Smerd + '/' + v.title)
                            }}

                        >
                            <ServiceCard disabled={false}
                                         {...v}
                            />
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
                                {...v}
                                disabled={true}
                                doRefresh={() => {
                                    console.log(123)
                                    window.location.reload()
                                }}
                            />
                        </div>)
                }
            </div>
        </div>
    )
}
