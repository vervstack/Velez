import {useState} from "react";
import {useParams, useNavigate} from "react-router-dom";
import {useQuery} from "@tanstack/react-query";
import cn from "classnames";

import {Smerd, SmerdStatus} from "@/app/api/velez";

import cls from "@/pages/smerd/SmerdPage.module.css";

import {FetchSmerd} from "@/processes/api/velez.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

export default function SmerdPage() {
    const params = useParams<Record<string, string>>();
    const navigate = useNavigate();
    const toaster = useToaster();
    const smerdName = params["name"] || "";

    const {data: smerd, isLoading} = useQuery({
        queryKey: ["smerd", smerdName],
        queryFn: () => FetchSmerd(smerdName).catch((e) => {
            toaster.catchGrpc(e);
            throw e;
        }),
        enabled: smerdName !== "",
        refetchInterval: 5000,
    });

    if (!smerdName) {
        return <div className={cls.SmerdPageContainer}><div className={cls.ErrorMessage}>No name provided.</div></div>;
    }

    if (isLoading) {
        return <div className={cls.SmerdPageContainer}><div className={cls.LoadingMessage}>Loading...</div></div>;
    }

    if (!smerd) {
        return <div className={cls.SmerdPageContainer}><div className={cls.ErrorMessage}>Container not found.</div></div>;
    }

    return (
        <div className={cls.SmerdPageContainer}>
            <div className={cls.BackLink} onClick={() => navigate("/")}>← Back</div>

            <div className={cls.Header}>
                <div className={cls.SmerdName}>{smerd.name}</div>
                <StatusBadge status={smerd.status}/>
            </div>

            <SmerdMetaSection smerd={smerd}/>

            {(smerd.ports && smerd.ports.length > 0) && (
                <InfoBox title="Ports">
                    {smerd.ports.map(function renderPort(p, i) {
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

            {(smerd.volumes && smerd.volumes.length > 0) && (
                <InfoBox title="Volumes">
                    {smerd.volumes.map(function renderVolume(v, i) {
                        return (
                            <div key={i} className={cls.MetaRow}>
                                <span className={cls.MetaLabel}>Volume</span>
                                <span className={cls.MetaValue}>{v.volumeName} → {v.containerPath}</span>
                            </div>
                        );
                    })}
                </InfoBox>
            )}

            <RawInspect smerd={smerd}/>
        </div>
    );
}

function SmerdMetaSection({smerd}: { smerd: Smerd }) {
    return (
        <InfoBox title="Details">
            {smerd.imageName && (
                <div className={cls.MetaRow}>
                    <span className={cls.MetaLabel}>Image</span>
                    <span className={cls.MetaValue}>{smerd.imageName}</span>
                </div>
            )}
            {smerd.uuid && (
                <div className={cls.MetaRow}>
                    <span className={cls.MetaLabel}>UUID</span>
                    <span className={cls.MetaValue}>{smerd.uuid}</span>
                </div>
            )}
            {smerd.createdAt && (
                <div className={cls.MetaRow}>
                    <span className={cls.MetaLabel}>Created at</span>
                    <span className={cls.MetaValue}>{formatTimestamp(smerd.createdAt)}</span>
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

function RawInspect({smerd}: { smerd: Smerd }) {
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
                    {JSON.stringify(smerd, null, 2)}
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

function formatTimestamp(ts: {seconds?: string | number; nanos?: number}): string {
    const seconds = Number(ts.seconds || 0);
    if (!seconds) return "—";
    return new Date(seconds * 1000).toLocaleString();
}
