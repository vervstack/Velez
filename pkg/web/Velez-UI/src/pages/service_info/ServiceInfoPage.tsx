import {useParams} from "react-router-dom";
import {useEffect, useState} from "react";

import {VervAppService} from "@vervstack/velez";

import cls from "@/pages/service_info/ServiceInfoPage.module.css";

import {GetServiceById, GetServiceByName} from "@/processes/api/service.ts";
import Header from "@/pages/service_info/parts/Header.tsx";
import Dialog from "@/components/complex/dialog/Dialog.tsx";
import DeployMenu from "@/pages/service_info/parts/DeployMenu.tsx";
import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";


export default function ServiceInfoPage() {
    const params = useParams<Record<string, string>>();

    const [serviceInfo, setServiceInfo] = useState<VervAppService | null>(null)
    const [dialogChild, setDialogChild] = useState<React.ReactNode | null>(null);

    const toaster = useToaster()

    function loadService() {
        const key = params['key'] || "";
        if (key == "") {
            toaster.bake({
                title: 'Can\'t load service',
                description: 'No service name or id provided',
                level: "Error"
            } as Toast)
        }

        const serviceId = Number(key)
        if (serviceId.toString() == key) {
            GetServiceById(serviceId.toString())
                .then(setServiceInfo)
        }
        GetServiceByName(key)
            .then(setServiceInfo)
    }

    useEffect(() => {
        loadService()
    }, []);

    function openDeployMenu() {
        if (!serviceInfo || !serviceInfo.id || !serviceInfo.name) return

        setDialogChild(
            <DeployMenu
                serviceId={serviceInfo.id}
                serviceName={serviceInfo.name}
            />)
    }

    if (!serviceInfo) {
        return (<div>Loading mock</div>)
    }

    return (
        <div className={cls.ServiceInfoPageContainer}>
            {serviceInfo.name &&
				<Header
					serviceName={serviceInfo.name}
					onClickDeploy={openDeployMenu}
				/>}

            <Dialog
                isOpen={dialogChild !== null}
                onClose={() => setDialogChild(null)}
                children={dialogChild}
            />
        </div>
    )
}
