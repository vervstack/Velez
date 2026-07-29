import {useEffect, useState} from 'react';

import {LoadingWrapper} from "@vervstack/chures";

import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

import {EnableStatefullCluster, VervPluginState} from '@/app/api/velez';
import {controlPlaneService} from "@/processes/api/control_plane.ts";
import {serviceService} from "@/processes/api/service.ts";
import {FetchNodeHardware} from "@/processes/api/velez.ts";
import {GetInitReq} from "@/processes/api/api.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import Button from "@/components/base/Button.tsx";
import Choice from "@/components/base/Choice.tsx";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";
import {useNavigate} from "react-router-dom";
import {Routes} from "@/app/router/Routes.ts";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";

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
    const [portNumber, setPortNumber] = useState('5432');
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

    function handlePortChange(e: React.ChangeEvent<HTMLInputElement>) {
        setPortNumber(e.target.value);
    }

    function handleEnable() {
        const payload: EnableStatefullCluster = {
            isExposePort: exposePort,
            exposeToPort: exposePort ? portNumber : undefined,
        };

        controlPlaneService.enableStatefullPgCluster(payload)
            .catch(toaster.catchGrpc)
    }

    function handleRestart() {
        serviceService.restartService(pl.serviceName)
            .catch(toaster.catchGrpc)
    }

    const portTooltip = isRunningInContainer
        ? undefined
        : "Running as a binary always exposes the port";

    let content;
    if (pl.state == VervPluginState.dead) {
        content = (
            <div className={cls.ActionSection}>
                <span>Service is down. Run it?</span>
                <Button onClick={handleRestart}>
                    Restart
                </Button>
            </div>
        )
    } else {
        content = (
            <div className={cls.ActionSection}>
                {pl.state == VervPluginState.disabled && (
                    <span>Service is disabled. Enable it?</span>
                )}

                <label
                    className={cls.CheckboxLabel}
                    data-tooltip-id={portTooltip ? "root-tooltip" : undefined}
                    data-tooltip-content={portTooltip}
                    data-tooltip-place="top"
                >
                    <Choice title={'Expose port'}
                            active={exposePort}
                            disabled={!isRunningInContainer}
                            onClick={() => setExposePort(!exposePort)}/>
                </label>
                {!isRunningInContainer && (
                    <span>Velez is running as a binary on this node, so the port is always exposed.</span>
                )}

                {exposePort && (
                    <div className={cls.InputGroup}>
                        <label className={cls.InputLabel}>Port number:</label>
                        <input
                            type="text"
                            className={cls.PortInput}
                            value={portNumber}
                            onChange={handlePortChange}
                        />
                    </div>
                )}
                <Button onClick={handleEnable}>
                    Enable
                </Button>
            </div>
        )
    }

    return (
        <LoadingWrapper isLoading={isLoadingHardware}>
            {content}
        </LoadingWrapper>
    );
}
