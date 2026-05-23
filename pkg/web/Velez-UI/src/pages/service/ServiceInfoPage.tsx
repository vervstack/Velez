import {useEffect, useState} from "react";
import {useNavigate, useParams} from "react-router-dom";
import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {serviceService} from "@/processes/api/service.ts";
import {GetServiceByNameQuery, ListDeploymentsByServiceNameQuery} from "@/processes/queries/services.ts";
import {ListSmerdsByServiceIdQuery} from "@/processes/queries/smerds.ts";

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
import {DeploymentStatus} from "@/app/api/velez";
import Button from "@/components/base/Button.tsx";

type ServiceTab = 'overview' | 'metrics' | 'instances' | 'history' | 'access';

interface TabInfo {
    id: ServiceTab;
    label: string;
}

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
            <BreadcrumbsBar serviceName={key}/>
            <ServicePageHeader
                serviceName={key}
                activeTab={activeTab}
                setActiveTab={setActiveTab}
            />

            <div className={cls.ServicePageContentWrapper}>
                {activeTab === 'overview' && (
                    <>
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
        </div>
    );
}


interface TabsProps {
    serviceName: string
    activeTab: ServiceTab
    setActiveTab: (s: ServiceTab) => void
}

function ServicePageHeader({serviceName, activeTab, setActiveTab}: TabsProps) {
    const serviceQuery = GetServiceByNameQuery(serviceName);
    const service = serviceQuery.data;

    if (!service || !service.name) return null;

    return (
        <div className={cls.ServicePageHeaderContainer}>
            <div className={cls.HeaderLeftWrapper}>
                <div className={cls.TabStripWrapper}>
                    {
                        TABS.map(t =>
                            <Tab
                                tab={t}
                                setActiveTab={setActiveTab}
                                isActive={activeTab === t.id} key={t.id}/>)}
                </div>

                <ActionsRow
                    serviceName={service.name}
                    serviceState={service.status || DeploymentStatus.DEPLOYMENT_STATUS_UNKNOWN}
                />
            </div>

            <div className={cls.HeaderRightWrapper}>
                <EnvSwitcher />
            </div>
        </div>
    );
}


interface BreadcrumbsBarProps {
    serviceName: string
}

function BreadcrumbsBar({serviceName}: BreadcrumbsBarProps) {
    const navigate = useNavigate();
    return (
        <div className={cls.BreadcrumbsBarContainer}>
            <span className={cls.BreadcrumbLink} onClick={() => navigate("/")}>services</span>
            <span className={cls.BreadcrumbSep}>/</span>
            <span className={cls.BreadcrumbCurrent}>{serviceName}</span>
        </div>
    );
}

interface ActionsRowProps {
    serviceName: string
    serviceState: DeploymentStatus
}

function ActionsRow({serviceName, serviceState}: ActionsRowProps) {
    const toaster = useToaster();
    const {OpenDialog, CloseDialog} = useDialog();
    const deploymentsQuery = ListDeploymentsByServiceNameQuery(serviceName);

    function handleStop() {

        serviceService.stopService(serviceName)
            .then(() => toaster.bake({
                title: "Service stopped",
                description: serviceName,
                level: "Info",
            } as Toast))
            .catch(toaster.catchGrpc)
            .finally(() => window.location.reload());
    }

    function handleRestart() {
        serviceService.restartService(serviceName)
            .then(() => toaster.bake({
                title: "Service restarted",
                description: serviceName,
                level: "Info",
            } as Toast))
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


    return (
        <div className={cls.HeaderActionsRow}>
            <Button
                onClick={handleStop}
                disabled={serviceState != DeploymentStatus.RUNNING}
            >
                ■ Stop
            </Button>

            <Button
                onClick={handleRestart}>
                {serviceState == DeploymentStatus.RUNNING ? '↺ Restart' : '▶ Start'}
            </Button>
            {/*TODO - remake deployment page than enable this*/}

            <span
                data-tooltip-id={"root-tooltip"}
                data-tooltip-content="Not available for now, deploy widget is not ready"
            >
                <Button
                    disabled={true}
                    onClick={openDeployMenu}>
                    + Deploy
                </Button>
            </span>
        </div>
    )
}


interface TabProps {
    tab: TabInfo
    isActive: boolean
    setActiveTab: (s: ServiceTab) => void
}

function Tab({tab, isActive, setActiveTab}: TabProps) {
    function onTabClick() {
        setActiveTab(tab.id);
    }

    return (
        <button
            key={tab.id}
            className={`${cls.TabButton} ${isActive ? cls.tabActive : ''}`}
            onClick={onTabClick}
        >
            {tab.label}
        </button>
    );
}
