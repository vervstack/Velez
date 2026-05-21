import {useState} from 'react';

import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

import {EnableStatefullCluster, VervPluginState} from '@/app/api/velez';
import {controlPlaneService} from "@/processes/api/control_plane.ts";
import {serviceService} from "@/processes/api/service.ts";
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

    const [exposePort, setExposePort] = useState(false);
    const [portNumber, setPortNumber] = useState('5432');

    function handlePortChange(e: React.ChangeEvent<HTMLInputElement>) {
        setPortNumber(e.target.value);
        console.log(pl)
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

    if (pl.state == VervPluginState.dead) {
        return (
            <div className={cls.ActionSection}>
                <span>Service is down. Run it?</span>
                <Button onClick={handleRestart}>
                    Restart
                </Button>
            </div>
        )
    }

    if (pl.state == VervPluginState.disabled) {
        return (
            <div className={cls.ActionSection}>
                <span>Service is disabled. Enable it?</span>
                <Button
                    onClick={handleEnable}
                >
                    Enable
                </Button>
            </div>
        )
    }

    return (
        <div className={cls.ActionSection}>
            <label className={cls.CheckboxLabel}>
                <Choice title={'Expose port'}
                        active={exposePort}
                        onClick={() => setExposePort(!exposePort)}/>
            </label>

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
    );
}
