import {useEffect, useState} from "react";

import styles from '@/pages/controlplane/ControlPlanePage.module.css';

import {ListServices} from "@/processes/api/control_plane.ts";

import {Service} from "@/model/services/Services";

import ServiceCard from "@/components/service/ServiceCard";
import useSettings from "@/app/settings/state.ts";
import Loader from "@/components/Loader.tsx";
import {useNavigate} from "react-router-dom";
import {Routes} from "@/app/router/Router.tsx";

export default function ControlPlanePage() {
    const [components, setComponents] =
        useState<Service[]>([])
    const navigate = useNavigate()


    const [isLoading, setIsLoading] = useState(true)

    const settings = useSettings();

    useEffect(() => {
        setIsLoading(true)
        ListServices(settings.initReq())
            .then(setComponents)
            .then(() => setIsLoading(false))
    }, [settings]);


    if (isLoading) {
        return (
            <div className={styles.ControlPlaneContainer}>
                <Loader/>
            </div>
        )
    }

    return (
        <div className={styles.ControlPlaneContainer}>
            {
                components
                    .map((v, idx) =>
                        <div
                            className={styles.ServiceCard}
                            key={v.title + idx}
                            onClick={() => {
                                navigate(Routes.Smerd + '/' + v.title)
                            }}
                        >
                            <ServiceCard {...v}/>
                        </div>)
            }
        </div>
    )
}
