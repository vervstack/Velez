import cls from "@/pages/service/widgets/GitlabRunnerActions.module.css";
import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import {serviceService} from "@/processes/api/service.ts";
import {GetServiceByNameQuery} from "@/processes/queries/services.ts";
import {ReregisterRunnerMutation} from "@/processes/queries/runners.ts";
import {DeploymentStatus} from "@/app/api/velez";
import Button from "@/components/base/Button.tsx";
import RunnerDropDialog from "@/dialogs/RunnerDropDialog/RunnerDropDialog.tsx";

interface Props {
    serviceName: string;
}

export default function GitlabRunnerActions({serviceName}: Props) {
    const toaster = useToaster();
    const {OpenDialog} = useDialog();
    const serviceQuery = GetServiceByNameQuery(serviceName);
    const reregisterRunner = ReregisterRunnerMutation();

    const serviceState = serviceQuery.data?.status || DeploymentStatus.DEPLOYMENT_STATUS_UNKNOWN;

    function handleStop() {
        const toast: Toast = {title: "Service stopped", description: serviceName, level: "Info"} as Toast;
        serviceService.stopService(serviceName)
            .then(() => toaster.bake(toast))
            .catch(toaster.catchGrpc)
            .finally(() => window.location.reload());
    }

    function handleRestart() {
        const toast: Toast = {title: "Service restarted", description: serviceName, level: "Info"} as Toast;
        serviceService.restartService(serviceName)
            .then(() => toaster.bake(toast))
            .catch(toaster.catchGrpc)
            .finally(() => window.location.reload());
    }

    function handleReregister() {
        reregisterRunner.mutate(serviceName, {
            onSuccess: () => {
                toaster.bake({title: "Runner reregistration started", description: serviceName, level: "Info"});
            },
            onError: toaster.catchGrpc,
        });
    }

    function openDropDialog() {
        OpenDialog(<RunnerDropDialog name={serviceName}/>);
    }

    return (
        <div className={cls.GitlabRunnerActionsContainer}>
            <Button
                onClick={handleStop}
                disabled={serviceState != DeploymentStatus.RUNNING}
            >
                ■ Stop
            </Button>

            <Button onClick={handleRestart}>
                {serviceState == DeploymentStatus.RUNNING ? "↺ Restart" : "▶ Start"}
            </Button>

            <Button onClick={handleReregister} disabled={reregisterRunner.isPending}>
                {reregisterRunner.isPending ? "Reregistering…" : "Rerun registration"}
            </Button>

            <Button variant="danger" onClick={openDropDialog}>
                ✕ Drop
            </Button>
        </div>
    );
}
