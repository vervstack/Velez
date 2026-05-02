import {useState, useCallback} from "react";
import {useParams, useNavigate} from "react-router-dom";
import {useQuery} from "@tanstack/react-query";

import {DeploymentStatus, DeploymentInfo} from "@/app/api/velez";

import cls from "@/pages/service/ServiceInfoPage.module.css";

import Dialog from "@/components/complex/dialog/Dialog.tsx";
import DeployMenu from "@/pages/service/parts/DeployMenu.tsx";
import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {FetchService, FetchDeployments} from "@/processes/api/service.ts";

const POLL_INTERVAL = 5000;

export default function ServiceInfoPage() {
    const params = useParams<Record<string, string>>();
    const navigate = useNavigate();
    const toaster = useToaster();
    const [dialogChild, setDialogChild] = useState<React.ReactNode | null>(null);

    const key = params["key"] || "";

    const serviceQuery = useQuery({
        queryKey: ["service", key],
        queryFn: () => FetchService(key).catch((e) => {
            toaster.catchGrpc(e);
            throw e;
        }),
        enabled: key !== "",
        refetchInterval: POLL_INTERVAL,
    });

    const service = serviceQuery.data;

    const deploymentsQuery = useQuery({
        queryKey: ["deployments", service?.id],
        queryFn: () => FetchDeployments(service!.id!).catch((e) => {
            toaster.catchGrpc(e);
            return {deployments: []};
        }),
        enabled: !!service?.id,
        refetchInterval: POLL_INTERVAL,
    });

    if (key === "") {
        return (
            <div className={cls.ServiceInfoPageContainer}>
                <div className={cls.ErrorMessage}>No service key provided.</div>
            </div>
        );
    }

    if (serviceQuery.isLoading) {
        return (
            <div className={cls.ServiceInfoPageContainer}>
                <div className={cls.LoadingMessage}>Loading service...</div>
            </div>
        );
    }

    if (!service) {
        return (
            <div className={cls.ServiceInfoPageContainer}>
                <div className={cls.ErrorMessage}>Service not found.</div>
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
            <DeployMenu serviceId={service.id} serviceName={service.name}/>
        );
    }

    const deployments = deploymentsQuery.data?.deployments || [];

    return (
        <div className={cls.ServiceInfoPageContainer}>
            <div className={cls.BackLink} onClick={() => navigate("/")}>← Back</div>

            <div className={cls.ServiceHeader}>
                <div className={cls.ServiceName}>{service.name}</div>
                <StatusBadge status={service.status}/>
            </div>

            <div className={cls.ActionBar}>
                <button className={cls.ActionButton} onClick={openDeployMenu}>
                    Deploy
                </button>
            </div>

            <div className={cls.MetaSection}>
                <MetaRow label="Service ID" value={service.id}/>
                <MetaRow label="Current deployment" value={service.currentDeploymentId}/>
                <MetaRow label="Status" value={service.status}/>
            </div>

            {deployments.length > 0 && (
                <DeploymentsSection
                    deployments={deployments}
                    currentDeploymentId={service.currentDeploymentId}
                />
            )}

            <Dialog
                isOpen={dialogChild !== null}
                onClose={() => setDialogChild(null)}
                children={dialogChild}
            />
        </div>
    );
}

function DeploymentsSection({deployments, currentDeploymentId}: {
    deployments: DeploymentInfo[];
    currentDeploymentId?: string;
}) {
    const [expandedId, setExpandedId] = useState<string | null>(null);

    const toggleExpand = useCallback(function toggleExpand(id: string) {
        setExpandedId(prev => prev === id ? null : id);
    }, []);

    return (
        <div className={cls.DeploymentsSection}>
            <div className={cls.SectionTitle}>Deployments</div>
            <div className={cls.DeploymentsList}>
                {deployments.map(function renderDeployment(dep) {
                    const isCurrent = dep.id === currentDeploymentId;
                    const isExpanded = expandedId === dep.id;

                    function onClick() {
                        toggleExpand(dep.id || "");
                    }

                    return (
                        <div key={dep.id} className={cls.DeploymentRowWrapper}>
                            <div className={`${cls.DeploymentRow} ${isCurrent ? cls.deploymentCurrent : ""}`} onClick={onClick}>
                                <span className={cls.DeployExpandIcon}>{isExpanded ? "▼" : "▶"}</span>
                                <span className={cls.DeployId}>{dep.id}</span>
                                {isCurrent && <span className={cls.CurrentBadge}>current</span>}
                                <span className={cls.DeployStatus}>
                                    <StatusBadge status={dep.status}/>
                                </span>
                            </div>
                            {isExpanded && (
                                <div className={cls.DeploymentDetail}>
                                    {dep.specId && (
                                        <div className={cls.MetaRow}>
                                            <span className={cls.MetaLabel}>Spec ID</span>
                                            <span className={cls.MetaValue}>{dep.specId}</span>
                                        </div>
                                    )}
                                    {dep.createdAt && (
                                        <div className={cls.MetaRow}>
                                            <span className={cls.MetaLabel}>Created at</span>
                                            <span className={cls.MetaValue}>{formatTimestamp(dep.createdAt)}</span>
                                        </div>
                                    )}
                                    <div className={cls.MetaRow}>
                                        <span className={cls.MetaLabel}>Raw</span>
                                        <pre className={cls.RawJson}>{JSON.stringify(dep, null, 2)}</pre>
                                    </div>
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

function formatTimestamp(ts: {seconds?: string | number; nanos?: number}): string {
    const seconds = Number(ts.seconds || 0);
    if (!seconds) return "—";
    return new Date(seconds * 1000).toLocaleString();
}

function MetaRow({label, value}: { label: string; value?: string }) {
    if (!value) return null;
    return (
        <div className={cls.MetaRow}>
            <span className={cls.MetaLabel}>{label}</span>
            <span className={cls.MetaValue}>{value}</span>
        </div>
    );
}

function StatusBadge({status}: { status?: DeploymentStatus | string }) {
    if (!status) return null;
    const isRunning = status === DeploymentStatus.RUNNING;
    const isError = status === DeploymentStatus.FAILED || status === DeploymentStatus.DELETED;
    let modClass = cls.statusUnknown;
    if (isRunning) modClass = cls.statusRunning;
    if (isError) modClass = cls.statusError;

    return <span className={`${cls.StatusBadge} ${modClass}`}>{status}</span>;
}
