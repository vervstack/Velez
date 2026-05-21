import {useNavigate} from 'react-router-dom';

import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

import {VervPluginState} from '@/app/api/velez';
import {Routes} from '@/app/router/Routes.ts';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import Button from '@/components/base/Button.tsx';
import {controlPlaneService} from '@/processes/api/control_plane.ts';
import {VervPlugin} from '@/model/services/VervPlugins.tsx';

export default function SimplePluginForm(pl: VervPlugin) {
    const toaster = useToaster();
    const {CloseDialog} = useDialog();
    const navigate = useNavigate();

    if (pl.state == VervPluginState.running) {
        navigate(Routes.Service + '/' + pl.serviceName);
        CloseDialog();
    }

    function handleEnable() {
        controlPlaneService.enablePlugin(pl.type).catch(toaster.catchGrpc);
    }

    if (pl.state == VervPluginState.dead) {
        return (
            <div className={cls.ActionSection}>
                <span>Service is down. Run it?</span>
                <Button>
                    Restart
                </Button>
            </div>
        );
    }

    if (pl.state == VervPluginState.disabled) {
        return (
            <div className={cls.ActionSection}>
                <span>Service is disabled. Enable it?</span>
                <Button onClick={handleEnable}>
                    Enable
                </Button>
            </div>
        );
    }

    return (
        <div className={cls.ActionSection}>
            <Button onClick={handleEnable}>
                Enable
            </Button>
        </div>
    );
}
