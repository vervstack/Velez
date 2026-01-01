import {useParams} from "react-router-dom";
import {useEffect, useState} from "react";

import {VervAppService} from "@vervstack/velez";

import cls from "@/pages/service_info/ServiceInfoPage.module.css";

import {useCredentialsStore} from "@/app/settings/creds.ts";
import {GetServiceById, GetServiceByName} from "@/processes/api/service.ts";
import Header from "@/pages/service_info/parts/Header.tsx";
import Dialog from "@/components/complex/dialog/Dialog.tsx";
import DeployMenu from "@/pages/service_info/parts/DeployMenu.tsx";


export default function ServiceInfoPage() {
    const params = useParams<Record<string, string>>();

    const [serviceInfo, setServiceInfo] = useState<VervAppService | null>(null)

    const [dialogChild, setDialogChild] = useState<React.ReactNode | null>(null);

    const credentialsStore = useCredentialsStore();

    function loadService() {
        const key = params['key'];
        if (!key) {
            // TODO redirect to new service
            throw 'No name or id provided'
        }

        const serviceId = Number(key)
        if (serviceId.toString() == key) {
            GetServiceById(credentialsStore.getInitReq(), serviceId.toString())
                .then(setServiceInfo)
            return
        }

        GetServiceByName(credentialsStore.getInitReq(), key)
            .then(setServiceInfo)
    }

    useEffect(() => {
        loadService()
    }, []);

    function openDeployMenu() {
        if (!serviceInfo || !serviceInfo.id) return

        setDialogChild(<DeployMenu/>)
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
