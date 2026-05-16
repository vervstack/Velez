import { useState, useEffect, useRef, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import cls from '@/pages/search/SearchPage.module.css';
import StatusDot from '@/components/base/StatusDot';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import { type ServiceCardData } from '@/components/service/ServiceCard';
import { FetchSmerds } from '@/processes/api/velez';
import { listSmerdsQuery } from '@/processes/queries/smerds';
import { mapSmerdToServiceCard } from '@/processes/mappings/smerds';
import { useToaster } from '@/app/hooks/toaster/Toaster';

type NodeInfo = { id: string; host: string; status: 'online' | 'degraded' | 'offline' };

export default function SearchPage() {
    const toaster = useToaster();
    const [query, setQuery] = useState('');
    const inputRef = useRef<HTMLInputElement>(null);

    const smerdsQuery = useQuery({
        ...listSmerdsQuery(),
        queryFn: () => FetchSmerds().catch((e) => { toaster.catchGrpc(e); return { smerds: [] }; }),
    });

    const services: ServiceCardData[] = (smerdsQuery.data?.smerds ?? []).map(mapSmerdToServiceCard);
    // TODO(T32): replace with ListNodes once backend API is available
    const nodes: NodeInfo[] = [];

    useEffect(function autoFocus() {
        inputRef.current?.focus();
    }, []);

    function handleQueryChange(e: React.ChangeEvent<HTMLInputElement>) {
        setQuery(e.target.value);
    }

    const { matchedServices, matchedNodes } = useMemo(function computeResults() {
        const q = query.trim().toLowerCase();
        if (!q) return { matchedServices: [], matchedNodes: [] };

        const matchedServices = services.filter(s =>
            s.name.toLowerCase().includes(q) ||
            s.image.toLowerCase().includes(q) ||
            s.node.id.toLowerCase().includes(q)
        );

        const matchedNodes = nodes.filter(n =>
            n.id.toLowerCase().includes(q) ||
            n.host.toLowerCase().includes(q)
        );

        return { matchedServices, matchedNodes };
    }, [services, nodes, query]);

    const hasResults = matchedServices.length > 0 || matchedNodes.length > 0;

    return (
        <div className={cls.SearchPageContainer}>
            <div className={cls.searchWrapper}>
                <span className={cls.searchIcon}>⌕</span>
                <input
                    ref={inputRef}
                    className={cls.searchInput}
                    value={query}
                    onChange={handleQueryChange}
                    placeholder="Search services, nodes…"
                />
            </div>

            {query && !hasResults && (
                <div className={cls.empty}>no results for "{query}"</div>
            )}

            {matchedServices.length > 0 && (
                <section className={cls.section}>
                    <div className={cls.sectionHeader}>
                        <span className={cls.sectionTitle}>Services</span>
                        <span className={cls.sectionCount}>{matchedServices.length}</span>
                    </div>
                    {matchedServices.map(function renderService(svc) {
                        return (
                            <div key={svc.name} className={cls.resultRow}>
                                <StatusDot status={svc.status} />
                                <span className={cls.resultName}>{svc.name}</span>
                                <div className={cls.resultChips}>
                                    <EnvChip env={svc.env} />
                                    {svc.incident && <IncidentChip />}
                                </div>
                                <span className={cls.resultMeta}>{svc.node.id}</span>
                            </div>
                        );
                    })}
                </section>
            )}

            {matchedNodes.length > 0 && (
                <section className={cls.section}>
                    <div className={cls.sectionHeader}>
                        <span className={cls.sectionTitle}>Nodes</span>
                        <span className={cls.sectionCount}>{matchedNodes.length}</span>
                    </div>
                    {matchedNodes.map(function renderNode(node) {
                        return (
                            <div key={node.id} className={cls.resultRow}>
                                <StatusDot status={node.status} />
                                <span className={cls.resultName}>{node.id}</span>
                                {node.status === 'degraded' && (
                                    <span className={cls.degradedLabel}>degraded</span>
                                )}
                                <span className={cls.resultMeta}>{node.host}</span>
                            </div>
                        );
                    })}
                </section>
            )}
        </div>
    );
}
