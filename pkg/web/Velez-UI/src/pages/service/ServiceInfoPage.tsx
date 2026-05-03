import {useState, useCallback} from "react";
import {useParams, useNavigate} from "react-router-dom";
import {useQuery} from "@tanstack/react-query";

import {DeploymentStatus, DeploymentInfo, Smerd} from "@/app/api/velez";

import cls from "@/pages/service/ServiceInfoPage.module.css";

import Dialog from "@/components/complex/dialog/Dialog.tsx";
import DeployMenu from "@/pages/service/parts/DeployMenu.tsx";
import {Toast, useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {FetchService, FetchDeployments, StopService, RestartService} from "@/processes/api/service.ts";
import {FetchSmerdsByServiceName} from "@/processes/api/velez.ts";

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

    const smerdsQuery = useQuery({
        queryKey: ["service-smerds", key],
        queryFn: () => FetchSmerdsByServiceName(key).catch((e) => {
            toaster.catchGrpc(e);
            return {smerds: []};
        }),
        enabled: key !== "",
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

    function openRemoveDialog() {
        if (!service?.name) {
            return;
        }
        const serviceName = service.name;

        function onCancel() {
            setDialogChild(null);
        }

        function onConfirmRemove() {
            const t: Toast = {
                title: "Service deletion not yet available",
                description: "",
                level: "Info",
            };
            toaster.bake(t);
            setDialogChild(null);
        }

        const removeDialogContent = (
            <div className={cls.RemoveDialogWrapper}>
                <div className={cls.RemoveDialogMessage}>
                    Remove service {serviceName}?
                </div>
                <div className={cls.RemoveDialogActions}>
                    <button className={cls.ActionButton} onClick={onCancel}>
                        Cancel
                    </button>
                    <button className={`${cls.ActionButton} ${cls.DangerButton}`} onClick={onConfirmRemove}>
                        Confirm Remove
                    </button>
                </div>
            </div>
        );
        setDialogChild(removeDialogContent);
    }

    function handleStop() {
        if (!service?.name) {
            return;
        }
        const serviceName = service.name;
        StopService(serviceName)
            .then(function onStopSuccess() {
                const t: Toast = {
                    title: "Service stopped",
                    description: serviceName,
                    level: "Info",
                };
                toaster.bake(t);
            })
            .catch(function onStopError(e) {
                toaster.catchGrpc(e);
            });
    }

    function handleRestart() {
        if (!service?.name) {
            return;
        }
        const serviceName = service.name;
        RestartService(serviceName)
            .then(function onRestartSuccess() {
                const t: Toast = {
                    title: "Service restarted",
                    description: serviceName,
                    level: "Info",
                };
                toaster.bake(t);
            })
            .catch(function onRestartError(e) {
                toaster.catchGrpc(e);
            });
    }

    const deployments = deploymentsQuery.data?.deployments || [];
    const currentSmerd = smerdsQuery.data?.smerds?.[0];

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
                <button className={cls.ActionButton} onClick={handleStop}>
                    Stop
                </button>
                <button className={cls.ActionButton} onClick={handleRestart}>
                    Restart
                </button>
                <button className={`${cls.ActionButton} ${cls.DangerButton}`} onClick={openRemoveDialog}>
                    Remove
                </button>
            </div>

            <div className={cls.MetaSection}>
                <MetaRow label="Service ID" value={service.id}/>
                <MetaRow label="Current deployment" value={service.currentDeploymentId}/>
                <MetaRow label="Status" value={service.status}/>
                {currentSmerd?.imageName && (
                    <MetaRow label="Image" value={currentSmerd.imageName}/>
                )}
            </div>

            {currentSmerd && (
                <SmerdMetaSection smerd={currentSmerd}/>
            )}

            <DeploymentsSection
                deployments={deployments}
                currentDeploymentId={service.currentDeploymentId}
            />

            <Dialog
                isOpen={dialogChild !== null}
                onClose={() => setDialogChild(null)}
                children={dialogChild}
            />
        </div>
    );
}

function SmerdMetaSection({smerd}: { smerd: Smerd }) {
    const hasEnv = smerd.env && Object.keys(smerd.env).length > 0;
    const hasPorts = smerd.ports && smerd.ports.length > 0;
    const hasVolumes = smerd.volumes && smerd.volumes.length > 0;

    if (!hasEnv && !hasPorts && !hasVolumes) {
        return null;
    }

    return (
        <div className={cls.SmerdMetaSection}>
            {hasPorts && (
                <div className={cls.SmerdMetaBlock}>
                    <div className={cls.SectionTitle}>Ports</div>
                    <div className={cls.PortsList}>
                        {smerd.ports!.map(function renderPort(port, i) {
                            const label = port.exposedTo
                                ? `${port.servicePortNumber} → ${port.exposedTo}`
                                : String(port.servicePortNumber);
                            return (
                                <span key={i} className={cls.PortTag}>
                                    {label}{port.protocol ? ` (${port.protocol})` : ""}
                                </span>
                            );
                        })}
                    </div>
                </div>
            )}
            {hasVolumes && (
                <div className={cls.SmerdMetaBlock}>
                    <div className={cls.SectionTitle}>Volumes</div>
                    <div className={cls.MetaSection}>
                        {smerd.volumes!.map(function renderVolume(vol, i) {
                            return (
                                <div key={i} className={cls.MetaRow}>
                                    <span className={cls.MetaLabel}>{vol.volumeName || "—"}</span>
                                    <span className={cls.MetaValue}>{vol.containerPath}</span>
                                </div>
                            );
                        })}
                    </div>
                </div>
            )}
            {hasEnv && (
                <div className={cls.SmerdMetaBlock}>
                    <div className={cls.SectionTitle}>Environment</div>
                    <div className={cls.MetaSection}>
                        {Object.entries(smerd.env!).map(function renderEnvRow([k, v]) {
                            return (
                                <div key={k} className={cls.MetaRow}>
                                    <span className={cls.MetaLabel}>{k}</span>
                                    <span className={cls.MetaValue}>{v}</span>
                                </div>
                            );
                        })}
                    </div>
                </div>
            )}
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
            {deployments.length === 0 ? (
                <div className={cls.DeploymentsEmptyState}>No deployments yet.</div>
            ) : (
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
            )}
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
