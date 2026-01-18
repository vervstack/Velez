import {useEffect, useState} from "react";

import {VervServiceType, EnableStatefullCluster, ServiceState} from "@vervstack/velez";

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {EnableService, EnableStatefullPgCluster, ListServices} from "@/processes/api/control_plane.ts";
import {Service} from "@/model/services/Services";
import ServiceCard from "@/components/service/ServiceCard";
import Loader from "@/components/Loader.tsx";

import {useCredentialsStore} from "@/app/settings/creds.ts";
import Dialog from "@/components/complex/dialog/Dialog.tsx";
import EnableStatefullMode from "@/widgets/services/EnableStatefullMode.tsx";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

export default function ControlPlanePage() {
    const [services, setServices] =
        useState<Service[]>([]);

    const [isLoading, setIsLoading] = useState(true)
    const [dialogChild, setDialogChild] = useState<React.ReactNode | undefined>();

    const credentialsStore = useCredentialsStore();
    const toaster = useToaster();

    const [isDeployRunning, setIsDeployRunning] = useState(false)

    useEffect(() => {
        setIsLoading(true)
        ListServices(credentialsStore.getInitReq())
            .then(setServices)
            .then(() => setIsLoading(false))
    }, []);


    if (isLoading) {
        return (
            <div className={cls.ControlPlaneContainer}>
                <Loader/>
            </div>
        )
    }

    function getDeployCallbackByType(serviceType: VervServiceType): (() => void) | undefined {
        switch (serviceType) {
            case VervServiceType.statefull_pg:
                return openDeployStatefullPgDiag
            case VervServiceType.makosh, VervServiceType.matreshka:
                return () => {
                    enableSimpleService(serviceType)
                }
            default:
                return
        }
    }

    function openDeployStatefullPgDiag(): void {
        setDialogChild(<EnableStatefullMode
            onDeploy={(r: EnableStatefullCluster) => {
                return EnableStatefullPgCluster(r, credentialsStore.getInitReq())
                    .catch(toaster.catchGrpc)
            }
            }/>);
    }

    function enableSimpleService(serviceType: VervServiceType.matreshka | VervServiceType.makosh) {
        if (isDeployRunning) {
            return
        }

        setIsDeployRunning(true)
        EnableService(serviceType, credentialsStore.getInitReq())
            .then(window.location.reload)
            .finally(() => {
                setIsDeployRunning(false)
            })
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
            <div className={cls.Content}>
                <div className={cls.ServicesBlock}>
                    {services
                        .filter(s => s.state == ServiceState.running)
                        .map((v, idx) =>
                            <div
                                className={cls.ServiceCardWrapper}
                                key={v.title + idx}
                                onClick={() => {
                                    // TODO when clicked will navigate to service's page
                                    // navigate(Routes.Smerd + '/' + v.title)
                                }}

                            >
                                <ServiceCard
                                    disabled={false}
                                    {...v}
                                />
                            </div>)
                    }
                </div>

                <div className={cls.ServicesBlock}>
                    {services
                        .filter(s => s.state !== ServiceState.running)
                        .map((v: Service, idx) =>
                            <div
                                className={cls.ServiceCardWrapper}
                                key={v.title + idx}
                            >
                                <ServiceCard
                                    {...v}
                                    disabled={true}
                                    onDeploy={getDeployCallbackByType(v.type)}
                                    isDeployRunning={isDeployRunning}
                                />
                            </div>)
                    }
                </div>
            </div>

            {/* TODO VERV-165 add animation for deployments*/}
            <Dialog
                isOpen={dialogChild !== undefined}
                onClose={() => setDialogChild(undefined)}
                children={dialogChild}
                blur={0.5}
            />
        </div>
    )
}

