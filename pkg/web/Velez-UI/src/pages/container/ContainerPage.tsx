import {useState, useEffect} from "react";
import {useParams, useNavigate} from "react-router-dom";
import cn from "classnames";

import {DockerContainer, SmerdStatus} from "@/app/api/velez";

import cls from "@/pages/container/ContainerPage.module.css";

import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {useBreadcrumbs} from "@/app/hooks/breadcrumbs/Breadcrumbs.ts";
import {useGetContainerQuery} from "@/processes/queries/containers.ts";
import {deriveContainerServicePresentation} from "@/processes/mappings/containers.ts";
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx";
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx";
import ServiceChip from "@/components/base/chips/ServiceChip.tsx";
import {Routes} from "@/app/router/Routes.ts";

export default function ContainerPage() {
    const params = useParams<Record<string, string>>();
    const navigate = useNavigate();
    const toaster = useToaster();
    const containerId = params["id"] || "";

    const {data: container, isLoading, isError, error, refetch} = useGetContainerQuery(
        containerId, containerId !== ""
    );
    useEffect(() => {
        if (error) toaster.catchGrpc(error);
    }, [error]);

    const setCrumbs = useBreadcrumbs((s) => s.setCrumbs);

    useEffect(function publishBreadcrumbs() {
        if (containerId === "") return;
        setCrumbs([
            {label: "Home", onClick: () => navigate("/")},
            {label: container?.name || containerId},
        ]);
        return function clearBreadcrumbs() {
            setCrumbs([]);
        };
    }, [containerId, container?.name, setCrumbs, navigate]);

    if (!containerId) {
        return <div className={cls.ContainerPageContainer}>
            <div className={cls.ErrorMessage}>No id provided.</div>
        </div>;
    }

    if (isLoading) {
        return <ContainerPageSkeleton/>;
    }

    if (isError) {
        return (
            <div className={cls.ContainerPageContainer}>
                <QueryErrorState message="Failed to load container." onRetry={refetch}/>
            </div>
        );
    }

    if (!container) {
        return <div className={cls.ContainerPageContainer}>
            <div className={cls.ErrorMessage}>Container not found.</div>
        </div>;
    }

    const servicePresentation = container.linkedServiceName
        ? deriveContainerServicePresentation(container)
        : null;

    return (
        <div className={cls.ContainerPageContainer}>
            <div className={cls.Header}>
                <div className={cls.ContainerNameWrapper}>
                    <div className={cls.ContainerName}>{container.name || container.id}</div>
                    <StatusBadge status={container.status}/>
                    {!container.isRegistered && (
                        <span className={cn(cls.StatusBadge, cls.statusUnknown)}>Unregistered</span>
                    )}
                </div>
                {servicePresentation && (
                    <ServiceChip
                        serviceName={servicePresentation.displayName}
                        icon={servicePresentation.icon}
                        onClick={() => navigate(Routes.Service + "/" + container.linkedServiceName)}
                    />
                )}
            </div>

            <ContainerMetaSection container={container}/>

            {(container.env && Object.keys(container.env).length > 0) && (
                <InfoBox title="Environment">
                    {Object.entries(container.env).map(function renderEnv([key, value]) {
                        return (
                            <div key={key} className={cls.MetaRow}>
                                <span className={cls.MetaLabel}>{key}</span>
                                <span className={cls.MetaValue}>{value}</span>
                            </div>
                        );
                    })}
                </InfoBox>
            )}

            {(container.networks && container.networks.length > 0) && (
                <InfoBox title="Networks">
                    {container.networks.map(function renderNetwork(n, i) {
                        const aliases = n.aliases && n.aliases.length > 0
                            ? ` (aliases: ${n.aliases.join(", ")})`
                            : "";
                        return (
                            <div key={i} className={cls.MetaRow}>
                                <span className={cls.MetaLabel}>{n.networkName}</span>
                                <span className={cls.MetaValue}>{n.ipAddress || "—"}{aliases}</span>
                            </div>
                        );
                    })}
                </InfoBox>
            )}

            {(container.mounts && container.mounts.length > 0) && (
                <InfoBox title="Mounts">
                    {container.mounts.map(function renderMount(m, i) {
                        const readOnly = m.readWrite ? "" : " (ro)";
                        return (
                            <div key={i} className={cls.MetaRow}>
                                <span className={cls.MetaLabel}>{m.type || "mount"}</span>
                                <span className={cls.MetaValue}>{m.source} → {m.destination}{readOnly}</span>
                            </div>
                        );
                    })}
                </InfoBox>
            )}

            {(container.ports && container.ports.length > 0) && (
                <InfoBox title="Ports">
                    {container.ports.map(function renderPort(p, i) {
                        return (
                            <div key={i} className={cls.MetaRow}>
                                <span className={cls.MetaLabel}>Port</span>
                                <span className={cls.MetaValue}>
                                    {p.servicePortNumber} → {p.exposedTo ?? "not exposed"} ({p.protocol})
                                </span>
                            </div>
                        );
                    })}
                </InfoBox>
            )}

            <RawInspect container={container}/>
        </div>
    );
}

function ContainerPageSkeleton() {
    return (
        <div className={cls.ContainerPageContainer}>
            <div className={cls.Header}>
                <SkeletonLoader shape="line" width="12rem" height="2rem"/>
                <SkeletonLoader shape="block" width="4rem" height="1.2rem"/>
            </div>
            <div className={cls.InfoBoxContainer}>
                <SkeletonLoader shape="line" width="6rem" height="0.8rem"/>
                <div className={cls.InfoBoxBody}>
                    <SkeletonLoader shape="line" width="60%" height="0.9rem"/>
                    <SkeletonLoader shape="line" width="40%" height="0.9rem"/>
                    <SkeletonLoader shape="line" width="50%" height="0.9rem"/>
                </div>
            </div>
        </div>
    );
}

function ContainerMetaSection({container}: { container: DockerContainer }) {
    return (
        <InfoBox title="Details">
            {container.imageName && (
                <div className={cls.MetaRow}>
                    <span className={cls.MetaLabel}>Image</span>
                    <span className={cls.MetaValue}>{container.imageName}</span>
                </div>
            )}
            {container.id && (
                <div className={cls.MetaRow}>
                    <span className={cls.MetaLabel}>Id</span>
                    <span className={cls.MetaValue}>{container.id}</span>
                </div>
            )}
            {container.createdAt && (
                <div className={cls.MetaRow}>
                    <span className={cls.MetaLabel}>Created at</span>
                    <span className={cls.MetaValue}>{formatTimestamp(container.createdAt)}</span>
                </div>
            )}
        </InfoBox>
    );
}

function InfoBox({title, children}: { title: string; children: React.ReactNode }) {
    return (
        <div className={cls.InfoBoxContainer}>
            <div className={cls.InfoBoxTitle}>{title}</div>
            <div className={cls.InfoBoxBody}>{children}</div>
        </div>
    );
}

function RawInspect({container}: { container: DockerContainer }) {
    const [open, setOpen] = useState(false);

    function toggle() {
        setOpen(prev => !prev);
    }

    return (
        <div className={cls.RawInspectContainer}>
            <button className={cls.RawInspectToggle} onClick={toggle}>
                {open ? "▼" : "▶"} Raw inspect
            </button>
            {open && (
                <pre className={cls.RawInspectBody}>
                    {JSON.stringify(container, null, 2)}
                </pre>
            )}
        </div>
    );
}

function StatusBadge({status}: { status?: SmerdStatus }) {
    if (!status) return null;
    const isRunning = status === SmerdStatus.running;
    const isError = status === SmerdStatus.exited || status === SmerdStatus.dead;
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

function formatTimestamp(ts: { seconds?: string | number; nanos?: number }): string {
    const seconds = Number(ts.seconds || 0);
    if (!seconds) return "—";
    return new Date(seconds * 1000).toLocaleString();
}
