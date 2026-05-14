import {useState} from "react";
import {useParams, useNavigate} from "react-router-dom";
import {useQuery} from "@tanstack/react-query";

import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {FetchService, FetchDeployments, StopService, RestartService} from "@/processes/api/service.ts";
import {FetchSmerdsByServiceName} from "@/processes/api/velez.ts";

import Dialog from "@/components/complex/dialog/Dialog.tsx";
import DeployMenu from "@/pages/service/parts/DeployMenu.tsx";

import EnvSwitcher from "@/widgets/service/EnvSwitcher/EnvSwitcher.tsx";
import ServiceHero from "@/widgets/service/ServiceHero/ServiceHero.tsx";
import ObservabilityTools from "@/widgets/service/ObservabilityTools/ObservabilityTools.tsx";
import ResourcesSection from "@/widgets/service/ResourcesSection/ResourcesSection.tsx";
import ServiceGraph from "@/widgets/service/ServiceGraph/ServiceGraph.tsx";
import DeploymentHistory from "@/widgets/service/DeploymentHistory/DeploymentHistory.tsx";
import Vervonomicon from "@/widgets/service/Vervonomicon/Vervonomicon.tsx";

import cls from "@/pages/service/ServiceInfoPage.module.css";

type ServiceTab = 'overview' | 'metrics' | 'instances' | 'history' | 'access';

const TABS: { id: ServiceTab; label: string }[] = [
    {id: 'overview', label: 'Overview'},
    {id: 'metrics', label: 'Metrics'},
    {id: 'instances', label: 'Instances'},
    {id: 'history', label: 'History'},
    {id: 'access', label: 'Access'},
];

export default function ServiceInfoPage() {
    const params = useParams<Record<string, string>>();
    const navigate = useNavigate();
    const toaster = useToaster();
    const [dialogChild, setDialogChild] = useState<React.ReactNode | null>(null);
    const [activeTab, setActiveTab] = useState<ServiceTab>('overview');

    const key = params["key"] || "";

    const serviceQuery = useQuery({
        queryKey: ["service", key],
        queryFn: () => FetchService(key).catch((e) => {
            toaster.catchGrpc(e);
            throw e;
        }),
        enabled: key !== "",
    });

    const service = serviceQuery.data;

    const deploymentsQuery = useQuery({
        queryKey: ["deployments", service?.id],
        queryFn: () => FetchDeployments(service!.id!).catch((e) => {
            toaster.catchGrpc(e);
            return {deployments: []};
        }),
        enabled: !!service?.id,
    });

    const smerdsQuery = useQuery({
        queryKey: ["service-smerds", key],
        queryFn: () => FetchSmerdsByServiceName(key).catch((e) => {
            toaster.catchGrpc(e);
            return {smerds: []};
        }),
        enabled: key !== "",
    });

    if (key === "") {
        return (
            <div className={cls.ServiceInfoPageContainer}>
                <div className={cls.StatusMessage}>No service key provided.</div>
            </div>
        );
    }

    if (serviceQuery.isLoading) {
        return (
            <div className={cls.ServiceInfoPageContainer}>
                <div className={cls.StatusMessage}>Loading service...</div>
            </div>
        );
    }

    if (!service) {
        return (
            <div className={cls.ServiceInfoPageContainer}>
                <div className={cls.StatusMessage}>Service not found.</div>
            </div>
        );
    }

    function openDeployMenu() {
        if (!service?.id || !service?.name) {
            toaster.bake({
                title: "Cannot deploy",
                description: "Service ID or name is missing",
                level: "Error",
            } as Toast);
            return;
        }
        setDialogChild(
            <DeployMenu
                serviceId={service.id}
                serviceName={service.name}
                onDeploymentCreated={() => {
                    setDialogChild(null);
                    deploymentsQuery.refetch();
                }}
            />
        );
    }

    function handleStop() {
        if (!service?.name) return;
        const serviceName = service.name;
        StopService(serviceName)
            .then(function onStopSuccess() {
                toaster.bake({title: "Service stopped", description: serviceName, level: "Info"} as Toast);
            })
            .catch(function onStopError(e) {
                toaster.catchGrpc(e);
            });
    }

    function handleRestart() {
        if (!service?.name) return;
        const serviceName = service.name;
        RestartService(serviceName)
            .then(function onRestartSuccess() {
                toaster.bake({title: "Service restarted", description: serviceName, level: "Info"} as Toast);
            })
            .catch(function onRestartError(e) {
                toaster.catchGrpc(e);
            });
    }

    const deployments = deploymentsQuery.data?.deployments || [];
    const currentSmerd = smerdsQuery.data?.smerds?.[0];

    return (
        <div className={cls.ServiceInfoPageContainer}>
            <div className={cls.ServicePageTabsWrapper}>
                <div className={cls.BreadcrumbWrapper}>
                    <span className={cls.BreadcrumbLink} onClick={() => navigate("/")}>services</span>
                    <span className={cls.BreadcrumbSep}>/</span>
                    <span className={cls.BreadcrumbCurrent}>{service.name}</span>
                </div>
                <div className={cls.TabStripWrapper}>
                    {TABS.map(function renderTab(tab) {
                        function onTabClick() {
                            setActiveTab(tab.id);
                        }
                        return (
                            <button
                                key={tab.id}
                                className={`${cls.TabButton} ${activeTab === tab.id ? cls.tabActive : ''}`}
                                onClick={onTabClick}
                            >
                                {tab.label}
                            </button>
                        );
                    })}
                </div>
                <div className={cls.TabActionsWrapper}>
                    <button className={cls.RestartButton} onClick={handleStop}>■ Stop</button>
                    <button className={cls.RestartButton} onClick={handleRestart}>↺ Restart</button>
                    <button className={cls.DeployButton} onClick={openDeployMenu}>+ Deploy</button>
                </div>
            </div>

            <div className={cls.ServicePageContentWrapper}>
                {activeTab === 'overview' && (
                    <>
                        <div className={cls.EnvSwitcherWrapper}>
                            <span className={cls.SectionLabel}>Environment</span>
                            <EnvSwitcher serviceId={service.id || ''} />
                        </div>
                        <ServiceHero
                            serviceId={service.id || ''}
                            serviceName={service.name || ''}
                            serviceStatus={service.status as string | undefined}
                            imageFromSmerd={currentSmerd?.imageName}
                        />
                        <div className={cls.ObservabilityWrapper}>
                            <span className={cls.SectionLabel}>Observability & Tools</span>
                            <ObservabilityTools serviceName={service.name || ''} />
                        </div>
                        <ResourcesSection serviceId={service.id || ''} />
                        <ServiceGraph serviceId={service.id || ''} serviceName={service.name || ''} />
                        <DeploymentHistory
                            deployments={deployments}
                            currentDeploymentId={service.currentDeploymentId}
                        />
                        <Vervonomicon serviceId={service.id || ''} serviceName={service.name || ''} />
                    </>
                )}
                {activeTab !== 'overview' && (
                    <div className={cls.ComingSoonWrapper}>
                        <span className={cls.ComingSoonText}>{activeTab} — coming soon</span>
                    </div>
                )}
            </div>

            <Dialog
                isOpen={dialogChild !== null}
                onClose={() => setDialogChild(null)}
                children={dialogChild}
            />
        </div>
    );
}
