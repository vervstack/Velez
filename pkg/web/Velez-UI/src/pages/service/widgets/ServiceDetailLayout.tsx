import {ReactNode, useEffect, useState} from "react";
import {useNavigate} from "react-router-dom";

import cls from "@/pages/service/widgets/ServiceDetailLayout.module.css";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {useBreadcrumbs} from "@/app/hooks/breadcrumbs/Breadcrumbs.ts";
import {GetServiceByNameQuery} from "@/processes/queries/services.ts";
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx";
import {ServiceTab} from "@/pages/service/widgets/tabs.ts";
import TabStrip from "@/pages/service/widgets/TabStrip.tsx";
import ServiceTagsStrip from "@/pages/service/widgets/ServiceTagsStrip.tsx";
import EnvSwitcher from "@/widgets/service/EnvSwitcher/EnvSwitcher.tsx";
import ServiceOverviewTab from "@/pages/service/widgets/ServiceOverviewTab.tsx";
import ServicePageSkeleton from "@/pages/service/widgets/ServicePageSkeleton.tsx";
import ServiceComingSoon from "@/pages/service/widgets/ServiceComingSoon.tsx";

interface Props {
    serviceName: string;
    headerActions?: ReactNode;
}

export default function ServiceDetailLayout({serviceName, headerActions}: Props) {
    const navigate = useNavigate();
    const toaster = useToaster();
    const setCrumbs = useBreadcrumbs((s) => s.setCrumbs);
    const [activeTab, setActiveTab] = useState<ServiceTab>('overview');

    const serviceQuery = GetServiceByNameQuery(serviceName);
    useEffect(function reportServiceError() {
        if (serviceQuery.error) toaster.catchGrpc(serviceQuery.error);
    }, [serviceQuery.error, toaster]);
    const service = serviceQuery.data;

    useEffect(function publishBreadcrumbs() {
        if (serviceName === "") return;
        setCrumbs([
            {label: "services", onClick: goToServices},
            {label: serviceName},
        ]);
        return function clearBreadcrumbs() {
            setCrumbs([]);
        };
    }, [serviceName, setCrumbs]);

    function goToServices() {
        navigate("/");
    }

    if (serviceName === "") {
        return (
            <div className={cls.ServiceDetailLayoutContainer}>
                <div className={cls.StatusMessage}>No service key provided.</div>
            </div>
        );
    }

    if (serviceQuery.isLoading) {
        return <ServicePageSkeleton/>;
    }

    if (serviceQuery.isError) {
        return (
            <div className={cls.ServiceDetailLayoutContainer}>
                <QueryErrorState message="Failed to load service." onRetry={serviceQuery.refetch}/>
            </div>
        );
    }

    if (!service || !service.name) {
        return (
            <div className={cls.ServiceDetailLayoutContainer}>
                <div className={cls.StatusMessage}>Service not found.</div>
            </div>
        );
    }

    return (
        <div className={cls.ServiceDetailLayoutContainer}>
            <div className={cls.ServicePageContentWrapper}>
                <div className={cls.PageActionsRow}>
                    <ServiceTagsStrip serviceName={serviceName}/>
                    {headerActions}
                </div>

                <div className={cls.TabsRow}>
                    <TabStrip activeTab={activeTab} setActiveTab={setActiveTab}/>
                    <EnvSwitcher serviceName={serviceName}/>
                </div>

                {activeTab === 'overview'
                    ? <ServiceOverviewTab serviceName={service.name}/>
                    : <ServiceComingSoon label={activeTab}/>}
            </div>
        </div>
    );
}
