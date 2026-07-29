import {useEffect, useState} from 'react';
import {useNavigate} from "react-router-dom";
import {LoadingWrapper} from "@vervstack/chures";

import {EnableStatefullCluster, VervPluginState} from '@/app/api/velez';
import {controlPlaneService} from "@/processes/api/control_plane.ts";
import {serviceService} from "@/processes/api/service.ts";
import {FetchNodeHardware} from "@/processes/api/velez.ts";
import {GetInitReq} from "@/processes/api/api.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import Button from "@/components/base/Button.tsx";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";
import {Routes} from "@/app/router/Routes.ts";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import StepsDialog from "@/dialogs/StepsDialog/StepsDialog.tsx";
import {StatefullPgContext} from "@/dialogs/PluginManageDialog/plugins/StatefullPgContext.ts";
import StatefullPgOverviewScreen
    from "@/dialogs/PluginManageDialog/plugins/screens/StatefullPgOverviewScreen.tsx";
import StatefullPgSettingsScreen
    from "@/dialogs/PluginManageDialog/plugins/screens/StatefullPgSettingsScreen.tsx";
import StatefullPgReviewScreen
    from "@/dialogs/PluginManageDialog/plugins/screens/StatefullPgReviewScreen.tsx";
import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

export default function StatefullPgPluginForm(pl: VervPlugin) {
    const toaster = useToaster();
    const {CloseDialog} = useDialog();

    const navigate = useNavigate();

    if (pl.state == VervPluginState.running) {
        navigate(Routes.Service + "/" + pl.serviceName)
        CloseDialog()
    }

    const [isRunningInContainer, setIsRunningInContainer] = useState(true);
    const [exposePort, setExposePort] = useState(false);
    const [portNumber] = useState('5432');
    const [isLoadingHardware, setIsLoadingHardware] = useState(true);

    useEffect(() => {
        FetchNodeHardware(GetInitReq())
            .then((res) => {
                if (!res.isRunningInContainer) {
                    setIsRunningInContainer(false);
                    setExposePort(true);
                }
            })
            .catch(toaster.catchGrpc)
            .finally(() => setIsLoadingHardware(false))
    }, []);

    function handleEnable(context: StatefullPgContext) {
        const payload: EnableStatefullCluster = {
            isExposePort: context.exposePort,
            exposeToPort: context.exposePort ? context.portNumber : undefined,
        };

        controlPlaneService.enableStatefullPgCluster(payload)
            .catch(toaster.catchGrpc)
    }

    function handleRestart() {
        serviceService.restartService(pl.serviceName)
            .catch(toaster.catchGrpc)
    }

    let content;
    if (pl.state == VervPluginState.dead) {
        content = (
            <div className={cls.ModalContainer}>
                <div className={cls.ModalContent}>
                    <div className={cls.ActionSection}>
                        <span>Service is down. Run it?</span>
                        <Button onClick={handleRestart}>
                            Restart
                        </Button>
                    </div>
                </div>
            </div>
        )
    } else {
        content = (
            <div className={cls.ActionSection}>
                <StepsDialog<StatefullPgContext>
                    steps={[
                        {name: 'overview', label: 'Overview', component: StatefullPgOverviewScreen},
                        {name: 'settings', label: 'Settings', component: StatefullPgSettingsScreen},
                        {name: 'review', label: 'Review', component: StatefullPgReviewScreen},
                    ]}
                    initialContext={{exposePort, portNumber, isRunningInContainer}}
                    onFinish={handleEnable}
                    icon="⛁"
                    eyebrow="Plugins"
                    eyebrowDetail="statefull-pg"
                />
            </div>
        )
    }

    return (
        <LoadingWrapper isLoading={isLoadingHardware}>
            {content}
        </LoadingWrapper>
    );
}
