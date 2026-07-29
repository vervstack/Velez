import {EnableStatefullCluster, VervPluginState} from '@/app/api/velez';
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {controlPlaneService} from "@/processes/api/control_plane.ts";
import {FetchNodeHardware} from "@/processes/api/velez.ts";
import {GetInitReq} from "@/processes/api/api.ts";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";
import {OpenWizardDialog} from "@/dialogs/StepsDialog/openWizardDialog.tsx";
import {StatefullPgContext} from "@/dialogs/PluginManageDialog/plugins/StatefullPgContext.ts";
import StatefullPgDeadPrompt from "@/dialogs/PluginManageDialog/plugins/StatefullPgDeadPrompt.tsx";
import StatefullPgOverviewScreen
    from "@/dialogs/PluginManageDialog/plugins/screens/StatefullPgOverviewScreen.tsx";
import StatefullPgSettingsScreen
    from "@/dialogs/PluginManageDialog/plugins/screens/StatefullPgSettingsScreen.tsx";
import StatefullPgReviewScreen
    from "@/dialogs/PluginManageDialog/plugins/screens/StatefullPgReviewScreen.tsx";

export function openStatefullPgDialog(plugin: VervPlugin): Promise<void> {
    if (plugin.state == VervPluginState.dead) {
        useDialog.getState().OpenDialog(<StatefullPgDeadPrompt plugin={plugin}/>);
        return Promise.resolve();
    }

    const steps = [
        {name: 'overview', label: 'Overview', component: StatefullPgOverviewScreen},
        {name: 'settings', label: 'Settings', component: StatefullPgSettingsScreen},
        {name: 'review', label: 'Review', component: StatefullPgReviewScreen},
    ];
    const header = {icon: '⛁', eyebrow: 'Plugins', eyebrowDetail: 'statefull-pg'};

    function handleEnable(context: StatefullPgContext) {
        const payload: EnableStatefullCluster = {
            isExposePort: context.exposePort,
            exposeToPort: context.exposePort ? context.portNumber : undefined,
        };

        controlPlaneService.enableStatefullPgCluster(payload)
            .catch(useToaster.getState().catchGrpc);
    }

    OpenWizardDialog<StatefullPgContext>({
        steps,
        initialContext: {exposePort: false, portNumber: '5432', isRunningInContainer: true},
        header,
        isLoading: true,
        onFinish: handleEnable,
    });

    return FetchNodeHardware(GetInitReq())
        .then((res) => {
            const isRunningInContainer = !!res.isRunningInContainer;
            const exposePort = !isRunningInContainer;

            OpenWizardDialog<StatefullPgContext>({
                steps,
                initialContext: {exposePort, portNumber: '5432', isRunningInContainer},
                header,
                onFinish: handleEnable,
            });
        })
        .catch(useToaster.getState().catchGrpc);
}
