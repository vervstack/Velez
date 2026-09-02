import { useMemo, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Toggle } from '@vervstack/chures';

import ServiceCard from '@/components/services/ServiceCard';
import SkeletonServiceCard from '@/components/service/SkeletonServiceCard';
import ServicesEmptyState from '@/pages/services/parts/ServicesEmptyState/ServicesEmptyState';
import { Routes } from '@/app/router/Routes';
import { useListServicesQuery } from '@/processes/queries/services';
import { useListSmerdsQuery } from '@/processes/queries/smerds';
import { mapServiceToListItem, ServiceListItem } from '@/processes/mappings/smerds';
import { useToaster } from '@/app/hooks/toaster/Toaster';
import { useDialog } from '@/app/hooks/dialog/Dialog.tsx';
import CreateAppDialog from '@/dialogs/CreateAppDialog/CreateAppDialog.tsx';
import Button from '@/components/base/Button.tsx';
import cls from '@/pages/services/ServicesPage.module.css';

const INCLUDE_INTERNAL_KEY = 'services.includeInternal';

function readIncludeInternal(): boolean {
    try {
        return localStorage.getItem(INCLUDE_INTERNAL_KEY) === 'true';
    } catch {
        return false;
    }
}

function writeIncludeInternal(value: boolean) {
    try {
        localStorage.setItem(INCLUDE_INTERNAL_KEY, String(value));
    } catch {
        // localStorage unavailable (private mode, blocked) — non-fatal.
    }
}

export default function ServicesPage() {
    const navigate = useNavigate();
    const toaster = useToaster();
    const {OpenDialog} = useDialog();

    const [search, setSearch] = useState('');
    const [includeInternal, setIncludeInternal] = useState(readIncludeInternal);

    const servicesQuery = useListServicesQuery(includeInternal);
    useEffect(() => {
        if (servicesQuery.error) toaster.catchGrpc(servicesQuery.error);
    }, [servicesQuery.error]);

    const smerdsQuery = useListSmerdsQuery();
    useEffect(() => {
        if (smerdsQuery.error) toaster.catchGrpc(smerdsQuery.error);
    }, [smerdsQuery.error]);

    const services: ServiceListItem[] = useMemo(function computeServices() {
        const smerds = smerdsQuery.data?.smerds ?? [];
        return (servicesQuery.data?.services ?? []).map(function toItem(s) {
            return mapServiceToListItem(s, smerds);
        });
    }, [servicesQuery.data, smerdsQuery.data]);

    const filtered = useMemo(function computeFiltered() {
        const q = search.trim().toLowerCase();
        if (!q) return services;
        return services.filter(
            (s) => s.name.toLowerCase().includes(q) || s.image.toLowerCase().includes(q)
        );
    }, [search, services]);

    function handleIncludeInternalChange(value: boolean) {
        setIncludeInternal(value);
        writeIncludeInternal(value);
    }

    function handleOpen(name: string) {
        navigate(Routes.Service + '/' + name);
    }

    function handleDeploy(name: string) {
        navigate(Routes.Service + '/' + name);
    }

    function handleSearchChange(e: React.ChangeEvent<HTMLInputElement>) {
        setSearch(e.target.value);
    }

    function handleCreate() {
        OpenDialog(<CreateAppDialog/>);
    }

    let gridContent: React.ReactNode;
    if (servicesQuery.isLoading) {
        gridContent = (
            <>
                <SkeletonServiceCard/>
                <SkeletonServiceCard/>
                <SkeletonServiceCard/>
            </>
        );
    } else if (services.length === 0) {
        gridContent = (
            <ServicesEmptyState includeInternal={includeInternal} onCreate={handleCreate}/>
        );
    } else {
        gridContent = filtered.map(function renderCard(service) {
            return (
                <ServiceCard
                    key={service.name}
                    service={service}
                    onOpen={handleOpen}
                    onDeploy={handleDeploy}
                />
            );
        });
    }

    return (
        <div className={cls.ServicesPageContainer}>
            <div className={cls.toolbar}>
                <input
                    className={cls.search}
                    placeholder="Filter services…"
                    value={search}
                    onChange={handleSearchChange}
                />
                <span className={cls.count}>{filtered.length} services</span>
                <div className={cls.toolbarRight}>
                    <Toggle
                        label="Show Verv internal"
                        labelPosition="right"
                        checked={includeInternal}
                        onChange={handleIncludeInternalChange}
                    />
                    <Button variant="primary" onClick={handleCreate}>
                        Create service
                    </Button>
                </div>
            </div>
            <div className={cls.grid}>
                {gridContent}
            </div>
        </div>
    );
}
