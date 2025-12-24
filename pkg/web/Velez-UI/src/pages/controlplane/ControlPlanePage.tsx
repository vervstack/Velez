import {useEffect, useState} from "react";

import {VervServiceType, EnableStatefullCluster} from "@vervstack/velez";

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {EnableService, EnableStatefullPgCluster, ListServices} from "@/processes/api/control_plane.ts";
import {Service} from "@/model/services/Services";
import ServiceCard from "@/components/service/ServiceCard";
import Loader from "@/components/Loader.tsx";

import {useCredentialsStore} from "@/app/settings/creds.ts";
import Dialog from "@/components/complex/dialog/Dialog.tsx";
import EnableStatefullMode from "@/widgets/services/EnableStatefullMode.tsx";

export default function ControlPlanePage() {
    const [activeComponents, setActiveComponents] =
        useState<Service[]>([]);

    const [inactiveComponents, setInactiveComponents] =
        useState<Service[]>([]);

    const [isLoading, setIsLoading] = useState(true)
    const [dialogChild, setDialogChild] = useState<React.ReactNode | undefined>();

    const credentialsStore = useCredentialsStore();

    const [isDeployRunning, setIsDeployRunning] = useState(false)

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
                EnableStatefullPgCluster(r, credentialsStore.getInitReq())
                    .then()
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
                                    onDeploy={getDeployCallbackByType(v.type)}
                                    isDeployRunning={isDeployRunning}
                                />
                            </div>)
                    }
                </div>
            </div>
            <Dialog
                isOpen={dialogChild !== undefined}
                closeDialog={() => setDialogChild(undefined)}
                children={dialogChild}
            />
        </div>
    )
}

