import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import {
    GetServiceByNameQuery,
    GetServiceProxyQuery,
    ListDeploymentsByServiceNameQuery,
} from "@/processes/queries/services.ts";
import Button from "@/components/base/Button.tsx";
import ServiceProxyDialog from "@/dialogs/ServiceProxyDialog/ServiceProxyDialog.tsx";

interface Props {
    serviceName: string;
}

export default function ServiceProxyButton({serviceName}: Props) {
    const {OpenDialog} = useDialog();
    const serviceQuery = GetServiceByNameQuery(serviceName);
    const deploymentsQuery = ListDeploymentsByServiceNameQuery(serviceName);
    const proxyQuery = GetServiceProxyQuery(serviceName);

    const currentProxyUrl = proxyQuery.data?.proxyUrl ?? "";

    function handleApplied() {
        serviceQuery.refetch();
        proxyQuery.refetch();
        deploymentsQuery.refetch();
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

    return (
        <Button onClick={openProxyDialog} disabled={proxyQuery.isLoading}>
            {currentProxyUrl ? "⇄ Proxy: on" : "⇄ Proxy"}
        </Button>
    );
}
