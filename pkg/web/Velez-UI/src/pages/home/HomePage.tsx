import {useNavigate} from "react-router-dom";
import {useState, useEffect, type MouseEvent} from "react";
import cn from "classnames";

import {ServiceBaseInfo, Smerd, SmerdStatus} from "@/app/api/velez";

import cls from "@/pages/home/Home.module.css";

import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import {Routes} from "@/app/router/Router.tsx";
import Button from "@/components/base/Button.tsx";
import SkeletonServiceCard from "@/components/service/SkeletonServiceCard.tsx";
import SkeletonSmerdRow from "@/components/smerd/SkeletonSmerdRow.tsx";
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx";
import {useListServicesQuery} from "@/processes/queries/services.ts";
import {useListSmerdsQuery} from "@/processes/queries/smerds.ts";
import {serviceService} from "@/processes/api/service.ts";
import RemoveServiceDialog from "@/dialogs/RemoveServiceDialog/RemoveServiceDialog.tsx";

export default function HomePage() {
    const toaster = useToaster();

    const servicesQuery = useListServicesQuery();
    useEffect(() => {
        if (servicesQuery.error) toaster.catchGrpc(servicesQuery.error);
    }, [servicesQuery.error]);

    const smerdsQuery = useListSmerdsQuery();
    useEffect(() => {
        if (smerdsQuery.error) toaster.catchGrpc(smerdsQuery.error);
    }, [smerdsQuery.error]);

    const services = servicesQuery.data?.services || [];
    const smerds = smerdsQuery.data?.smerds || [];
    const isLoadingServices = servicesQuery.isLoading;
    const isLoadingSmerds = smerdsQuery.isLoading;
    const isServicesError = servicesQuery.isError;
    const isSmerdsError = smerdsQuery.isError;

    const hasContent = services.length > 0 || smerds.length > 0;

    return (
        <div className={cls.HomeContainer}>
            {isLoadingServices || isLoadingSmerds ? (
                <div className={cls.DashboardWrapper}>
                    {isLoadingServices && <ServicesSectionSkeleton />}
                    {isLoadingSmerds && <SmerdsSectionSkeleton />}
                </div>
            ) : hasContent || isServicesError || isSmerdsError ? (
                <div className={cls.DashboardWrapper}>
                    <ServicesSection
                        services={services}
                        smerds={smerds}
                        isError={isServicesError}
                        onRetry={servicesQuery.refetch}
                    />
                    <SmerdsSection
                        smerds={smerds}
                        isError={isSmerdsError}
                        onRetry={smerdsQuery.refetch}
                    />
                </div>
            ) : (
                <EmptyState/>
            )}
        </div>
    );
}

function formatLastDeployed(ts?: {seconds?: string | number; nanos?: number}): string {
    if (!ts?.seconds) return "";
    const seconds = Number(ts.seconds);
    if (!seconds) return "";
    const date = new Date(seconds * 1000);
    const diffMs = Date.now() - date.getTime();
    const diffDays = Math.floor(diffMs / 86400000);
    if (diffDays === 0) {
        const diffHours = Math.floor(diffMs / 3600000);
        if (diffHours === 0) {
            const diffMinutes = Math.floor(diffMs / 60000);
            return diffMinutes <= 1 ? "just now" : `${diffMinutes}m ago`;
        }
        return `${diffHours}h ago`;
    }
    if (diffDays === 1) return "yesterday";
    if (diffDays < 30) return `${diffDays}d ago`;
    return date.toLocaleDateString();
}

interface ServicesSectionProps {
    services: ServiceBaseInfo[];
    smerds: Smerd[];
    isError: boolean;
    onRetry: () => void;
}

function ServicesSection({services, smerds, isError, onRetry}: ServicesSectionProps) {
    const navigate = useNavigate();
    const {OpenDialog, CloseDialog} = useDialog();
    const [query, setQuery] = useState("");

    if (isError) {
        return (
            <section className={cls.Section}>
                <h2 className={cls.SectionTitle}>Services</h2>
                <QueryErrorState message="Failed to load services." onRetry={onRetry}/>
            </section>
        );
    }

    if (services.length === 0) {
        return null;
    }

    const filteredServices = services.filter(s => s.name?.toLowerCase().includes(query.toLowerCase()));

    return (
        <section className={cls.Section}>
            <h2 className={cls.SectionTitle}>Services</h2>
            <input
                type="text"
                className={cls.SearchInput}
                placeholder="Search services..."
                value={query}
                onChange={function onQueryChange(e) { setQuery(e.target.value); }}
            />
            {filteredServices.length === 0 && services.length > 0 ? (
                <div className={cls.EmptyFilter}>No services match your search.</div>
            ) : (
                <div className={cls.CardGrid}>
                    {filteredServices.map(function renderServiceCard(service: ServiceBaseInfo) {
                        function onClick() {
                            navigate(Routes.Service + "/" + service.name);
                        }

                        function openRemoveDialog(e?: MouseEvent) {
                            e?.stopPropagation();
                            OpenDialog(
                                <RemoveServiceDialog
                                    serviceName={service.name || ""}
                                    onCancel={CloseDialog}
                                    onRemoved={() => {
                                        CloseDialog();
                                        onRetry();
                                    }}
                                />
                            );
                        }

                        const relatedSmerd = smerds.find(
                            (s) => s.name === service.name || s.name?.startsWith(service.name + "-")
                        );
                        const imageLabel = relatedSmerd?.imageName || null;

                        const lastDeployed = formatLastDeployed(service.lastDeployedAt);

                        return (
                            <div
                                key={service.name}
                                className={cls.ServiceCardContainer}
                                onClick={onClick}
                            >
                                <div className={cls.CardName}>{service.name}</div>
                                {imageLabel && (
                                    <div className={cls.CardImage}>{imageLabel}</div>
                                )}
                                {service.repo && (
                                    <div className={cls.CardRepo}>{service.repo}</div>
                                )}
                                {lastDeployed && (
                                    <div className={cls.CardLastDeployed}>{lastDeployed}</div>
                                )}
                                <div className={cls.CardStatus}>
                                    <StatusBadge status={service.status || "unknown"}/>
                                </div>
                                <ServiceCardActions
                                    serviceName={service.name || ""}
                                    onChanged={onRetry}
                                    onRemoveClick={openRemoveDialog}
                                />
                            </div>
                        );
                    })}
                </div>
            )}
        </section>
    );
}

interface ServiceCardActionsProps {
    serviceName: string;
    onChanged: () => void;
    onRemoveClick: (e?: MouseEvent) => void;
}

function ServiceCardActions({serviceName, onChanged, onRemoveClick}: ServiceCardActionsProps) {
    const toaster = useToaster();

    function handleStop(e?: MouseEvent) {
        e?.stopPropagation();
        serviceService.stopService(serviceName)
            .then(() => {
                toaster.bake({
                    title: "Service stopped",
                    description: serviceName,
                    level: "Info",
                } as Toast);
                onChanged();
            })
            .catch(toaster.catchGrpc);
    }

    function handleRestart(e?: MouseEvent) {
        e?.stopPropagation();
        serviceService.restartService(serviceName)
            .then(() => {
                toaster.bake({
                    title: "Service restarted",
                    description: serviceName,
                    level: "Info",
                } as Toast);
                onChanged();
            })
            .catch(toaster.catchGrpc);
    }

    return (
        <div className={cls.CardActions}>
            <Button sm onClick={handleStop}>Stop</Button>
            <Button sm onClick={handleRestart}>Restart</Button>
            <Button sm variant="danger" onClick={onRemoveClick}>Remove</Button>
        </div>
    );
}


interface SmerdsSectionProps {
    smerds: Smerd[];
    isError: boolean;
    onRetry: () => void;
}

function SmerdsSection({smerds, isError, onRetry}: SmerdsSectionProps) {
    const navigate = useNavigate();

    return (
        <section className={cls.Section}>
            <h2 className={cls.SectionTitle}>Containers</h2>
            {isError ? (
                <QueryErrorState message="Failed to load containers." onRetry={onRetry}/>
            ) : smerds.length === 0 ? (
                <div className={cls.EmptyFilter}>No containers running.</div>
            ) : (
                <div className={cls.SmerdList}>
                    {smerds.map(function renderSmerdRow(smerd: Smerd) {
                        function onClick() {
                            navigate(Routes.Smerd + "/" + smerd.name);
                        }

                        return (
                            <div
                                key={smerd.uuid || smerd.name}
                                className={cls.SmerdRowContainer}
                                onClick={onClick}
                            >
                                <div className={cls.SmerdDot}>
                                    <SmerdStatusDot status={smerd.status}/>
                                </div>
                                <div className={cls.SmerdName}>{smerd.name}</div>
                                <div className={cls.SmerdImage}>{smerd.imageName}</div>
                                <div className={cls.SmerdStatusLabel}>
                                    <StatusBadge status={smerd.status || "unknown"}/>
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </section>
    );
}

function EmptyState() {
    const navigate = useNavigate();

    function onCreateService() {
        navigate(Routes.NewVervService);
    }

    return (
        <div className={cls.EmptyStateContainer}>
            <div className={cls.EmptyStateIcon}>◻</div>
            <div className={cls.EmptyStateMessage}>No services or containers on this cluster yet.</div>
            <div className={cls.EmptyStateAction}>
                <Button title="Create a service" onClick={onCreateService}/>
            </div>
        </div>
    );
}

function ServicesSectionSkeleton() {
    return (
        <section className={cls.Section}>
            <h2 className={cls.SectionTitle}>Services</h2>
            <div className={cls.CardGrid}>
                <SkeletonServiceCard />
                <SkeletonServiceCard />
                <SkeletonServiceCard />
            </div>
        </section>
    );
}

function SmerdsSectionSkeleton() {
    return (
        <section className={cls.Section}>
            <h2 className={cls.SectionTitle}>Containers</h2>
            <div className={cls.SmerdList}>
                <SkeletonSmerdRow />
                <SkeletonSmerdRow />
                <SkeletonSmerdRow />
            </div>
        </section>
    );
}

function SmerdStatusDot({status}: { status?: SmerdStatus }) {
    const isRunning = status === SmerdStatus.running;
    const isError = status === SmerdStatus.exited || status === SmerdStatus.dead;
    return (
        <div className={cn(cls.StatusDot, {
            [cls.statusRunning]: isRunning,
            [cls.statusError]: isError,
            [cls.statusUnknown]: !isRunning && !isError,
        })}/>
    );
}

function StatusBadge({status}: { status: string }) {
    const isRunning = status === SmerdStatus.running || status === "RUNNING";
    const isError = status === SmerdStatus.exited || status === SmerdStatus.dead || status === "FAILED";
    return (
        <span className={cn(cls.StatusBadge, {
            [cls.statusRunning]: isRunning,
            [cls.statusError]: isError,
            [cls.statusUnknown]: !isRunning && !isError,
        })}>
            {status}
        </span>
    );
}
