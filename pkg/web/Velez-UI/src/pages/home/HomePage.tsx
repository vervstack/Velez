import {useQuery} from "@tanstack/react-query";
import {useNavigate} from "react-router-dom";
import cn from "classnames";

import {ServiceBaseInfo, Smerd, SmerdStatus} from "@/app/api/velez";

import cls from "@/pages/home/Home.module.css";

import {ListServices} from "@/processes/api/service.ts";
import {FetchSmerds} from "@/processes/api/velez.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {Routes} from "@/app/router/Router.tsx";
import Button from "@/components/base/Button.tsx";

const POLL_INTERVAL = 5000;
const LIST_REQ = {paging: {limit: '50', offset: '0'}};

export default function HomePage() {
    const toaster = useToaster();

    const servicesQuery = useQuery({
        queryKey: ["services"],
        queryFn: () => ListServices(LIST_REQ).catch((e) => {
            toaster.catchGrpc(e);
            return {services: []};
        }),
        refetchInterval: POLL_INTERVAL,
    });

    const smerdsQuery = useQuery({
        queryKey: ["smerds"],
        queryFn: () => FetchSmerds().catch((e) => {
            toaster.catchGrpc(e);
            return {smerds: []};
        }),
        refetchInterval: POLL_INTERVAL,
    });

    const services = servicesQuery.data?.services || [];
    const smerds = smerdsQuery.data?.smerds || [];
    const isLoading = servicesQuery.isLoading && smerdsQuery.isLoading;

    if (isLoading) {
        return <div className={cls.HomeContainer}><div className={cls.LoadingMessage}>Loading...</div></div>;
    }

    const hasContent = services.length > 0 || smerds.length > 0;

    return (
        <div className={cls.HomeContainer}>
            {hasContent ? (
                <div className={cls.DashboardWrapper}>
                    <ServicesSection services={services}/>
                    <SmerdsSection smerds={smerds}/>
                </div>
            ) : (
                <EmptyState/>
            )}
        </div>
    );
}

function ServicesSection({services}: { services: ServiceBaseInfo[] }) {
    const navigate = useNavigate();

    if (services.length === 0) {
        return null;
    }

    return (
        <section className={cls.Section}>
            <h2 className={cls.SectionTitle}>Services</h2>
            <div className={cls.CardGrid}>
                {services.map(function renderServiceCard(service: ServiceBaseInfo) {
                    function onClick() {
                        navigate(Routes.Service + "/" + service.name);
                    }

                    return (
                        <div
                            key={service.name}
                            className={cls.ServiceCardContainer}
                            onClick={onClick}
                        >
                            <div className={cls.CardName}>{service.name}</div>
                            <div className={cls.CardStatus}>
                                <StatusBadge status="unknown"/>
                            </div>
                        </div>
                    );
                })}
            </div>
        </section>
    );
}

function SmerdsSection({smerds}: { smerds: Smerd[] }) {
    const navigate = useNavigate();

    if (smerds.length === 0) {
        return null;
    }

    return (
        <section className={cls.Section}>
            <h2 className={cls.SectionTitle}>Containers</h2>
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
