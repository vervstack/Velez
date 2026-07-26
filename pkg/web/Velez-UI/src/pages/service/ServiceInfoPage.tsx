import {useEffect, useState} from "react";
import {useNavigate, useParams} from "react-router-dom";
import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {serviceService} from "@/processes/api/service.ts";
import {GetServiceByNameQuery, ListDeploymentsByServiceNameQuery} from "@/processes/queries/services.ts";
import {ListSmerdsByServiceIdQuery} from "@/processes/queries/smerds.ts";
import {getSmerdTags} from "@/processes/mappings/smerds.ts";

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
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx";
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx";
import BreadcrumbsBar from "@/components/complex/BreadcrumbsBar/BreadcrumbsBar.tsx";
import TagChip from "@/components/base/chips/TagChip.tsx";
import RemoveServiceDialog from "@/dialogs/RemoveServiceDialog/RemoveServiceDialog.tsx";

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
    const navigate = useNavigate();
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

    if (key === "") {
        return (
            <div className={cls.ServiceInfoPageContainer}>
                <div className={cls.StatusMessage}>No service key provided.</div>
            </div>
        );
    }

    if (serviceQuery.isLoading) {
        return <ServicePageSkeleton/>;
    }

    if (serviceQuery.isError) {
        return (
            <div className={cls.ServiceInfoPageContainer}>
                <QueryErrorState message="Failed to load service." onRetry={serviceQuery.refetch}/>
            </div>
        );
    }

    if (!service || !service.name) {
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
            <BreadcrumbsBar crumbs={[
                {label: "services", onClick: () => navigate("/")},
                {label: key},
            ]}/>
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


function ServicePageSkeleton() {
    return (
        <div className={cls.ServiceInfoPageContainer}>
            <div className={cls.ServicePageContentWrapper}>
                <SkeletonLoader shape="block" width="100%" height="6rem"/>
                <SkeletonLoader shape="block" width="100%" height="8rem"/>
                <SkeletonLoader shape="block" width="100%" height="10rem"/>
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
                <EnvSwitcher serviceName={serviceName} />
                <ServiceTagsStrip serviceName={serviceName} />
            </div>
        </div>
    );
}


interface ActionsRowProps {
    serviceName: string
    serviceState: DeploymentStatus
}

function ActionsRow({serviceName, serviceState}: ActionsRowProps) {
    const toaster = useToaster();
    const navigate = useNavigate();
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

            <Button onClick={openDeployMenu}>
                + Deploy
            </Button>

            <Button variant="danger" onClick={openRemoveDialog}>
                ✕ Remove
            </Button>
        </div>
    )
}


interface ServiceTagsStripProps {
    serviceName: string
}

function ServiceTagsStrip({serviceName}: ServiceTagsStripProps) {
    const smerdsQuery = ListSmerdsByServiceIdQuery(serviceName);
    const currentSmerd = smerdsQuery.data?.smerds?.[0];
    const tags = getSmerdTags(currentSmerd);

    if (tags.length === 0) return null;

    return (
        <div className={cls.TagsStrip}>
            {tags.map(function renderTag(tag) {
                return <TagChip key={tag.key} tagKey={tag.key} value={tag.value}/>;
            })}
        </div>
    );
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
