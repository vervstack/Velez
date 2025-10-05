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
    }, [initReq]);

    function updateComponentsList(components: Service[]) {
        const componentsList: Service[] = [
            {
                tittle: 'Velez',
                icon: <VelezIcon/>,
                externalLink: ""
            },
        ]

        componentsList.push(...components)
        setComponents(componentsList)
    }

    return (
        <div className={styles.ControlPlane}>
            {
                components
                    .map((v) =>
                        <div
                            className={styles.ServiceCard}
                            key={v.tittle}
                        >
                            <ServiceCard {...v}/>
                        </div>)
            }
        </div>
    )
}
