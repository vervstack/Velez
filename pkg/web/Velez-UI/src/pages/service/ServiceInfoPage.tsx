import {useState, useEffect} from "react";
import {useParams, useNavigate} from "react-router-dom";
import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {serviceService} from "@/processes/api/service.ts";
import {GetServiceByNameQuery, ListDeploymentsByServiceNameQuery} from "@/processes/queries/services.ts";
import {ListSmerdsByServiceIdQuery} from "@/processes/queries/smerds.ts";

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
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";

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
    const toaster = useToaster();
    const [dialogChild, setDialogChild] = useState<React.ReactNode | null>(null);
    const [activeTab, setActiveTab] = useState<ServiceTab>('overview');

    const key = params["key"] || "";

    const serviceQuery = GetServiceByNameQuery(key);
    useEffect(() => {
        if (serviceQuery.error) toaster.catchGrpc(serviceQuery.error);
    }, [serviceQuery.error, toaster]);
    const service = serviceQuery.data;


    const deploymentsQuery = ListDeploymentsByServiceNameQuery(service?.name || "");

    const smerdsQuery = ListSmerdsByServiceIdQuery(service?.name || "");

    if (!service || !service.name) return null;

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


    const deployments = deploymentsQuery.data?.deployments || [];
    const currentSmerd = smerdsQuery.data?.smerds?.[0];

    return (
        <div className={cls.ServiceInfoPageContainer}>
            <Tabs
                serviceName={key}
                activeTab={activeTab}
                setActiveTab={setActiveTab}
            />

            <div className={cls.ServicePageContentWrapper}>
                {activeTab === 'overview' && (
                    <>
                        <div className={cls.EnvSwitcherWrapper}>
                            <span className={cls.SectionLabel}>Environment</span>
                            <EnvSwitcher serviceName={service.name}/>
                        </div>
                        <ServiceHero
                            serviceName={service.name}
                            serviceStatus={service.status as string | undefined}
                            imageFromSmerd={currentSmerd?.imageName}
                        />
                        <div className={cls.ObservabilityWrapper}>
                            <span className={cls.SectionLabel}>Observability & Tools</span>
                            <ObservabilityTools serviceName={service.name}/>
                        </div>
                        <ResourcesSection serviceName={service.name}/>
                        <ServiceGraph serviceName={service.name}/>
                        <DeploymentHistory
                            deployments={deployments}
                            currentDeploymentId={service.currentDeploymentId}
                        />
                        <Vervonomicon serviceName={service.name}/>
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


interface TabsProps {
    serviceName: string

    activeTab: ServiceTab
    setActiveTab: (s: ServiceTab) => void
}

function Tabs(
    {
        serviceName,
        activeTab, setActiveTab
    }: TabsProps) {
    const navigate = useNavigate();
    const toaster = useToaster();
    const {OpenDialog, CloseDialog} = useDialog();


    const serviceQuery = GetServiceByNameQuery(serviceName);
    const service = serviceQuery.data

    if (!service || !service.name) return null;

    const deploymentsQuery = ListDeploymentsByServiceNameQuery(service.name)

    function openDeployMenu() {
        if (!service?.name) {
            toaster.bake({
                title: "Cannot deploy",
                description: "Service name is missing",
                level: "Error",
            } as Toast);
            return;
        }
        OpenDialog(
            <DeployMenu
                serviceName={service.name}
                onDeploymentCreated={() => {
                    CloseDialog();
                    deploymentsQuery.refetch();
                }}
            />
        );
    }

    function handleStop() {
        console.debug(service)

        if (!service?.name) return;

        const serviceName = service.name;

        serviceService.stopService(serviceName)
            .then(() => toaster
                .bake({
                    title: "Service stopped",
                    description: serviceName,
                    level: "Info"
                } as Toast))
            .catch(toaster.catchGrpc);
    }

    function handleRestart() {
        if (!service?.name) return;
        const serviceName = service.name;
        serviceService.restartService(serviceName)
            .then(function onRestartSuccess() {
                toaster.bake({title: "Service restarted", description: serviceName, level: "Info"} as Toast);
            })
            .catch(function onRestartError(e) {
                toaster.catchGrpc(e);
            });
    }

    return (
        <div className={cls.ServicePageTabsWrapper}>
            <div className={cls.BreadcrumbWrapper}>
                <span className={cls.BreadcrumbLink} onClick={() => navigate("/")}>services</span>
                <span className={cls.BreadcrumbSep}>/</span>
                <span className={cls.BreadcrumbCurrent}>{serviceName}</span>
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
    )
}
