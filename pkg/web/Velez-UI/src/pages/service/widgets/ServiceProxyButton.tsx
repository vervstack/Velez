import {useEffect, useState} from "react";
import cn from "classnames";

import cls from "@/pages/service/widgets/ServiceProxyButton.module.css";
import {TaskStatusStatus} from "@/app/api/velez";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import {useServiceUpgrade} from "@/app/hooks/serviceUpgrade/ServiceUpgrade.ts";
import {useUpgradeProgress} from "@/app/hooks/serviceUpgrade/UpgradeProgress.ts";
import {
    GetServiceByNameQuery,
    GetServiceProxyQuery,
    ListDeploymentsByServiceNameQuery,
} from "@/processes/queries/services.ts";
import Button from "@/components/base/Button.tsx";
import TypewriterText from "@/components/complex/TypewriterText/TypewriterText.tsx";
import ServiceProxyDialog from "@/dialogs/ServiceProxyDialog/ServiceProxyDialog.tsx";

const TERMINAL_LINGER_MS = 2500;

interface Props {
    serviceName: string;
}

export default function ServiceProxyButton({serviceName}: Props) {
    const {OpenDialog} = useDialog();
    const watch = useServiceUpgrade((state) => state.watch);
    const taskStatus = useServiceUpgrade((state) => state.statusByService[serviceName]?.status);
    const {isInFlight, percent, stepLabel, isFailed} = useUpgradeProgress(serviceName);
    const [isLingerOver, setIsLingerOver] = useState(false);
    const serviceQuery = GetServiceByNameQuery(serviceName);
    const deploymentsQuery = ListDeploymentsByServiceNameQuery(serviceName);
    const proxyQuery = GetServiceProxyQuery(serviceName);

    const currentProxyUrl = proxyQuery.data?.proxyUrl ?? "";
    const isTerminal = taskStatus === TaskStatusStatus.DONE || taskStatus === TaskStatusStatus.FAILED;
    const isLineVisible = isInFlight || (isTerminal && !isLingerOver);

    useEffect(() => {
        watch(serviceName);
    }, [serviceName]);

    useEffect(() => {
        if (!isTerminal) {
            return;
        }

        refetchServiceState();

        setIsLingerOver(false);
        const timer = setTimeout(() => setIsLingerOver(true), TERMINAL_LINGER_MS);

        return () => clearTimeout(timer);
    }, [taskStatus]);

    function refetchServiceState() {
        serviceQuery.refetch();
        proxyQuery.refetch();
        deploymentsQuery.refetch();
    }

    function handleApplied() {
        watch(serviceName);
        refetchServiceState();
    }

    function openProxyDialog() {
        OpenDialog(
            <ServiceProxyDialog
                serviceName={serviceName}
                currentProxyUrl={currentProxyUrl}
                currentBypassHosts={proxyQuery.data?.proxyBypassHosts ?? []}
                onApplied={handleApplied}
            />
        );
    }

    function buttonLabel(): string {
        if (isInFlight) {
            return `⇄ Proxy ${percent}%`;
        }

        return currentProxyUrl ? "⇄ Proxy: on" : "⇄ Proxy";
    }

    return (
        <div className={cls.ServiceProxyButtonContainer}>
            <Button onClick={openProxyDialog} disabled={proxyQuery.isLoading || isInFlight}>
                {buttonLabel()}
            </Button>

            <div className={cn(cls.StepLineWrapper, {[cls.IsVisible]: isLineVisible, [cls.IsFailed]: isFailed})}>
                <TypewriterText text={stepLabel}/>
            </div>
        </div>
    );
}
