import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import cls from '@/pages/apps/AppsPage.module.css';
import AppCard from '@/components/apps/AppCard';
import { Routes } from '@/app/router/Routes';
import { FetchSmerds } from '@/processes/api/velez';
import { mapSmerdToAppData } from '@/processes/mappings/smerds';
import { useToaster } from '@/app/hooks/toaster/Toaster';
import { AppData } from '@/processes/mappings/smerds';

export default function AppsPage() {
    const navigate = useNavigate();
    const [search, setSearch] = useState('');
    const toaster = useToaster();

    const smerdsQuery = useQuery({
        queryKey: ['smerds'],
        queryFn: () =>
            FetchSmerds().catch((e) => {
                toaster.catchGrpc(e);
                return { smerds: [] };
            }),
        refetchInterval: 5000,
    });

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
