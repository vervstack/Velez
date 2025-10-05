import {useEffect, useState} from "react";

import styles from './ControlPlane.module.css';

import VelezIcon from "@/assets/icons/velez/VelezIcon";
import {ListServices} from "@/processes/api/control_plane.ts";

import {Service} from "@/model/services/Services";

import ServiceCard from "@/components/service/ServiceCard";
import useSettings from "@/app/settings/state.ts";

export default function ControlPageWidget() {
    const [components, setComponents] =
        useState<Service[]>([])

    const {initReq} = useSettings();

    useEffect(() => {
        ListServices(initReq())
            .then(updateComponentsList)
    }, []);

    function updateComponentsList(components: Service[]) {
        const componentsList: Service[] = [
            {
                title: 'Velez',
                icon: <VelezIcon/>,
                webLink: "/"
            },
        ]

        componentsList.push(...components)
        setComponents(componentsList)
    }

    return (
        <div className={styles.ControlPlaneContainer}>
            {
                components
                    .map((v, idx) =>
                        <div
                            className={styles.ServiceCard}
                            key={v.title + idx}
                        >
                            <ServiceCard {...v}/>
                        </div>)
            }
        </div>
    )
}
