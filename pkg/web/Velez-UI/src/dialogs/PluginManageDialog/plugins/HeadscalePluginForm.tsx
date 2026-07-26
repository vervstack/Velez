import {useState} from 'react';

import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

import {EnableHeadscaleServer, VervPluginState} from '@/app/api/velez';
import {controlPlaneService} from "@/processes/api/control_plane.ts";
import {serviceService} from "@/processes/api/service.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import Button from "@/components/base/Button.tsx";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";
import {useNavigate} from "react-router-dom";
import {Routes} from "@/app/router/Routes.ts";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";

export default function HeadscalePluginForm(pl: VervPlugin) {
    const toaster = useToaster();
    const {CloseDialog} = useDialog();

    const navigate = useNavigate();

    if (pl.state == VervPluginState.running) {
        navigate(Routes.Service + "/" + pl.serviceName)
        CloseDialog()
    }

    const [headscaleMode, setHeadscaleMode] = useState<'deploy' | 'external'>('deploy');
    const [customPort, setCustomPort] = useState('');
    const [customImage, setCustomImage] = useState('');
    const [externalUrl, setExternalUrl] = useState('');
    const [externalToken, setExternalToken] = useState('');

    function handleSetDeploy() {
        setHeadscaleMode('deploy');
    }

    function handleSetExternal() {
        setHeadscaleMode('external');
    }

    function handleCustomPortChange(e: React.ChangeEvent<HTMLInputElement>) {
        setCustomPort(e.target.value);
    }

    function handleCustomImageChange(e: React.ChangeEvent<HTMLInputElement>) {
        setCustomImage(e.target.value);
    }

    function handleExternalUrlChange(e: React.ChangeEvent<HTMLInputElement>) {
        setExternalUrl(e.target.value);
    }

    function handleExternalTokenChange(e: React.ChangeEvent<HTMLInputElement>) {
        setExternalToken(e.target.value);
    }

    function handleEnable() {
        let payload: EnableHeadscaleServer;

        if (headscaleMode === 'deploy') {
            payload = {
                deployConfig: {
                    customPort: customPort ? parseInt(customPort, 10) : undefined,
                    customImage: customImage || undefined,
                },
            };
        } else {
            payload = {
                externalConnect: {
                    url: externalUrl,
                    token: externalToken,
                },
            };
        }

        controlPlaneService.enableHeadscaleServer(payload)
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
                <div className={cls.RadioGroup}>
                    <label className={cls.RadioLabel}>
                        <input
                            type="radio"
                            value="deploy"
                            checked={headscaleMode === 'deploy'}
                            onChange={handleSetDeploy}
                        />
                        <span>Deploy new</span>
                    </label>
                    <label className={cls.RadioLabel}>
                        <input
                            type="radio"
                            value="external"
                            checked={headscaleMode === 'external'}
                            onChange={handleSetExternal}
                        />
                        <span>Connect to external</span>
                    </label>
                </div>

                {headscaleMode === 'deploy' && (
                    <div className={cls.ConfigSection}>
                        <div className={cls.InputGroup}>
                            <label className={cls.InputLabel}>Custom port (optional):</label>
                            <input
                                type="text"
                                className={cls.ConfigInput}
                                value={customPort}
                                onChange={handleCustomPortChange}
                                placeholder="e.g., 8443"
                            />
                        </div>
                        <div className={cls.InputGroup}>
                            <label className={cls.InputLabel}>Custom image (optional):</label>
                            <input
                                type="text"
                                className={cls.ConfigInput}
                                value={customImage}
                                onChange={handleCustomImageChange}
                                placeholder="e.g., juanfont/headscale:v0.22.0"
                            />
                        </div>
                    </div>
                )}

                {headscaleMode === 'external' && (
                    <div className={cls.ConfigSection}>
                        <div className={cls.InputGroup}>
                            <label className={cls.InputLabel}>Server URL:</label>
                            <input
                                type="text"
                                className={cls.ConfigInput}
                                value={externalUrl}
                                onChange={handleExternalUrlChange}
                                placeholder="https://headscale.example.com"
                            />
                        </div>
                        <div className={cls.InputGroup}>
                            <label className={cls.InputLabel}>API Token:</label>
                            <input
                                type="text"
                                className={cls.ConfigInput}
                                value={externalToken}
                                onChange={handleExternalTokenChange}
                                placeholder="Your API token"
                            />
                        </div>
                    </div>
                )}

                <Button onClick={handleEnable}>
                    Enable
                </Button>
            </div>
        )
    }

    return (
        <div className={cls.ActionSection}>
            <Button onClick={handleEnable}>
                Enable
            </Button>
        </div>
    );
}
