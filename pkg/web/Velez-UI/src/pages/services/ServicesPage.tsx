import { useMemo, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Toggle } from '@vervstack/chures';

import ServiceCard from '@/components/services/ServiceCard';
import ContainerSection from '@/pages/services/parts/ContainerSection/ContainerSection.tsx';
import RegistrationFilterSelect from '@/pages/services/parts/RegistrationFilter/RegistrationFilter.tsx';
import {groupContainers} from '@/pages/services/processes/groupContainers.ts';
import {
    containersOfEntry,
    partitionByRegistration,
    RegistrationFilter,
} from '@/pages/services/processes/partitionByRegistration.ts';
import SkeletonServiceCard from '@/components/service/SkeletonServiceCard';
import ServicesEmptyState from '@/pages/services/parts/ServicesEmptyState/ServicesEmptyState';
import { Routes } from '@/app/router/Routes';
import { useListServicesQuery } from '@/processes/queries/services';
import { useListSmerdsQuery } from '@/processes/queries/smerds';
import { useListContainersQuery } from '@/processes/queries/containers';
import { parseContainerSearch } from '@/processes/queries/parseContainerSearch';
import { ContainerFilter, ContainerFilterField } from '@/app/api/velez';
import { mapServiceToListItem, ServiceListItem } from '@/processes/mappings/smerds';
import { useToaster } from '@/app/hooks/toaster/Toaster';
import { useDialog } from '@/app/hooks/dialog/Dialog.tsx';
import CreateServiceDialog from '@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx';
import Button from '@/components/base/Button.tsx';
import cls from '@/pages/services/ServicesPage.module.css';

const INCLUDE_INTERNAL_KEY = 'services.includeInternal';
const SHOW_ALL_CONTAINERS_KEY = 'services.showAllContainers';

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

function readShowAllContainers(): boolean {
    try {
        return localStorage.getItem(SHOW_ALL_CONTAINERS_KEY) === 'true';
    } catch {
        return false;
    }
}

function writeShowAllContainers(value: boolean) {
    try {
        localStorage.setItem(SHOW_ALL_CONTAINERS_KEY, String(value));
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
    const [showAllContainers, setShowAllContainers] = useState(readShowAllContainers);
    const [registrationFilter, setRegistrationFilter] = useState<RegistrationFilter>('all');

    const servicesQuery = useListServicesQuery(includeInternal);
    useEffect(() => {
        if (servicesQuery.error) toaster.catchGrpc(servicesQuery.error);
    }, [servicesQuery.error]);

    const smerdsQuery = useListSmerdsQuery();
    useEffect(() => {
        if (smerdsQuery.error) toaster.catchGrpc(smerdsQuery.error);
    }, [smerdsQuery.error]);

    const containerQuery = useMemo(function computeContainerQuery() {
        return parseContainerSearch(search);
    }, [search]);

    const containerFilters: ContainerFilter[] | undefined = useMemo(function computeContainerFilters() {
        if (containerQuery.mode !== 'service' || containerQuery.value === '') return undefined;
        return [{field: ContainerFilterField.service, value: containerQuery.value}];
    }, [containerQuery]);

    const containersQuery = useListContainersQuery(containerFilters);
    useEffect(() => {
        if (containersQuery.error) toaster.catchGrpc(containersQuery.error);
    }, [containersQuery.error]);

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
            (s) => s.name.toLowerCase().includes(q)
                || s.displayName.toLowerCase().includes(q)
                || s.image.toLowerCase().includes(q)
        );
    }, [search, services]);

    const filteredContainers = useMemo(function computeFilteredContainers() {
        const containers = containersQuery.data?.containers ?? [];
        if (containerQuery.mode === 'service') return containers;
        const q = containerQuery.value.toLowerCase();
        if (!q) return containers;
        return containers.filter(
            (c) => (c.name ?? '').toLowerCase().includes(q) || (c.imageName ?? '').toLowerCase().includes(q)
        );
    }, [containerQuery, containersQuery.data]);

    const containerLayout = useMemo(function computeContainerLayout() {
        return groupContainers(filteredContainers);
    }, [filteredContainers]);

    const containerSections = useMemo(function computeContainerSections() {
        return partitionByRegistration(containerLayout, registrationFilter);
    }, [containerLayout, registrationFilter]);

    const visibleContainerCount = useMemo(function computeVisibleContainerCount() {
        return [...containerSections.registered, ...containerSections.unregistered, ...containerSections.awaitingEnd]
            .reduce((total, entry) => total + containersOfEntry(entry).length, 0);
    }, [containerSections]);

    function handleIncludeInternalChange(value: boolean) {
        setIncludeInternal(value);
        writeIncludeInternal(value);
    }

    function handleShowAllContainersChange(value: boolean) {
        setShowAllContainers(value);
        writeShowAllContainers(value);
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
        OpenDialog(<CreateServiceDialog/>);
    }

    function handleOpenContainer(id: string) {
        navigate(Routes.Container + '/' + id);
    }

    function handleFilterByService(serviceName: string) {
        setSearch(`service: ${serviceName}`);
    }

    let gridContent: React.ReactNode;
    let countLabel: string;
    if (showAllContainers) {
        countLabel = `${visibleContainerCount} containers`;
        if (containersQuery.isLoading) {
            gridContent = (
                <>
                    <SkeletonServiceCard/>
                    <SkeletonServiceCard/>
                    <SkeletonServiceCard/>
                </>
            );
        } else if (visibleContainerCount === 0) {
            gridContent = <div className={cls.containersEmpty}>No containers on this node.</div>;
        } else {
            gridContent = (
                <>
                    {containerSections.awaitingEnd.length > 0 && (
                        <ContainerSection
                            title="Awaiting onboarding end"
                            entries={containerSections.awaitingEnd}
                            onOpen={handleOpenContainer}
                            onFilterByService={handleFilterByService}
                        />
                    )}
                    {containerSections.registered.length > 0 && (
                        <ContainerSection
                            title="Registered"
                            entries={containerSections.registered}
                            onOpen={handleOpenContainer}
                            onFilterByService={handleFilterByService}
                        />
                    )}
                    {containerSections.unregistered.length > 0 && (
                        <ContainerSection
                            title="Not onboarded"
                            entries={containerSections.unregistered}
                            onOpen={handleOpenContainer}
                            onFilterByService={handleFilterByService}
                        />
                    )}
                </>
            );
        }
    } else {
        countLabel = `${filtered.length} services`;
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
                <span className={cls.count}>{countLabel}</span>
                <div className={cls.toolbarRight}>
                    <Toggle
                        label="Show VervStack Services"
                        checked={includeInternal}
                        onChange={handleIncludeInternalChange}
                    />
                    {showAllContainers && (
                        <RegistrationFilterSelect value={registrationFilter} onChange={setRegistrationFilter}/>
                    )}
                    <Toggle
                        label="Show all containers"
                        checked={showAllContainers}
                        onChange={handleShowAllContainersChange}
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
