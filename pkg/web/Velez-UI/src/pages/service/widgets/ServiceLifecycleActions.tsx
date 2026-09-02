import {useNavigate} from "react-router-dom";

import cls from "@/pages/service/widgets/ServiceLifecycleActions.module.css";
import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import {serviceService} from "@/processes/api/service.ts";
import {GetServiceByNameQuery, ListDeploymentsByServiceNameQuery} from "@/processes/queries/services.ts";
import {DeploymentStatus} from "@/app/api/velez";
import Button from "@/components/base/Button.tsx";
import RemoveServiceDialog from "@/dialogs/RemoveServiceDialog/RemoveServiceDialog.tsx";
import DeployMenu from "@/pages/service/widgets/DeployMenu.tsx";

interface Props {
    serviceName: string;
}

export default function ServiceLifecycleActions({serviceName}: Props) {
    const toaster = useToaster();
    const navigate = useNavigate();
    const {OpenDialog, CloseDialog} = useDialog();
    const serviceQuery = GetServiceByNameQuery(serviceName);
    const deploymentsQuery = ListDeploymentsByServiceNameQuery(serviceName);

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

    function openDeployMenu() {
        OpenDialog(
            <DeployMenu
                serviceName={serviceName}
                onDeploymentCreated={() => {
                    CloseDialog();
                    deploymentsQuery.refetch();
                }}
            />
        );
    }

    function openRemoveDialog() {
        OpenDialog(
            <RemoveServiceDialog
                serviceName={serviceName}
                onCancel={CloseDialog}
                onRemoved={() => {
                    CloseDialog();
                    navigate("/");
                }}
            />
        );
    }

    return (
        <div className={cls.ServiceLifecycleActionsContainer}>
            <Button
                onClick={handleStop}
                disabled={serviceState != DeploymentStatus.RUNNING}
            >
                ■ Stop
            </Button>

            <Button onClick={handleRestart}>
                {serviceState == DeploymentStatus.RUNNING ? '↺ Restart' : '▶ Start'}
            </Button>

            <Button onClick={openDeployMenu}>
                + Deploy
            </Button>

            <Button variant="danger" onClick={openRemoveDialog}>
                ✕ Remove
            </Button>
        </div>
    );
}
