import { useMemo, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import cls from '@/pages/apps/AppsPage.module.css';
import AppCard from '@/components/apps/AppCard';
import { Routes } from '@/app/router/Routes';
import { useListSmerdsQuery } from '@/processes/queries/smerds';
import { mapSmerdToAppData } from '@/processes/mappings/smerds';
import { useToaster } from '@/app/hooks/toaster/Toaster';
import { AppData } from '@/processes/mappings/smerds';

export default function AppsPage() {
    const navigate = useNavigate();
    const [search, setSearch] = useState('');
    const toaster = useToaster();

    const smerdsQuery = useListSmerdsQuery();
    useEffect(() => {
        if (smerdsQuery.error) toaster.catchGrpc(smerdsQuery.error);
    }, [smerdsQuery.error]);

    const apps: AppData[] = (smerdsQuery.data?.smerds ?? []).map(mapSmerdToAppData);

    const filtered = useMemo(function computeFiltered() {
        const q = search.trim().toLowerCase();
        if (!q) return apps;
        return apps.filter(
            (a) =>
                a.name.toLowerCase().includes(q) ||
                a.image.toLowerCase().includes(q)
        );
    }, [search, apps]);

    function handleOpen(name: string) {
        navigate(Routes.Service + '/' + name);
    }

    function handleDeploy(name: string) {
        console.log('deploy', name);
    }

    function handleSearchChange(e: React.ChangeEvent<HTMLInputElement>) {
        setSearch(e.target.value);
    }

    return (
        <div className={cls.AppsPageContainer}>
            <div className={cls.toolbar}>
                <input
                    className={cls.search}
                    placeholder="Filter apps…"
                    value={search}
                    onChange={handleSearchChange}
                />
                <span className={cls.count}>{filtered.length} apps</span>
            </div>
            <div className={cls.grid}>
                {filtered.map(function renderCard(app) {
                    return (
                        <AppCard
                            key={app.name}
                            app={app}
                            onOpen={handleOpen}
                            onDeploy={handleDeploy}
                        />
                    );
                })}
            </div>
        </div>
    );
}
