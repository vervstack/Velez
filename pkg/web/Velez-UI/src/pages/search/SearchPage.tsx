import { useState, useEffect, useRef, useMemo } from 'react';
import cls from '@/pages/search/SearchPage.module.css';
import StatusDot from '@/components/base/StatusDot';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import { type ServiceCardData } from '@/components/service/ServiceCard';

// Mock data — replace with React Query calls
const MOCK_SERVICES: ServiceCardData[] = [
    { name: 'matreshka-be', image: 'godverv/matreshka:latest', status: 'running',  cpu: 12, mem: 180,  uptime: '14d 3h', restarts: 0,  env: 'prod',  incident: false, releaseFrozen: false, node: { id: 'node01', host: '192.168.1.10', status: 'online'   } },
    { name: 'api-gateway',  image: 'internal/gateway:v2.1',    status: 'running',  cpu: 33, mem: 512,  uptime: '6d 4h',  restarts: 0,  env: 'prod',  incident: true,  releaseFrozen: false, node: { id: 'node02', host: '192.168.1.42', status: 'online'   } },
    { name: 'postgres-main',image: 'postgres:16-alpine',        status: 'degraded', cpu: 89, mem: 1800, uptime: '3d 7h',  restarts: 12, env: 'prod',  incident: true,  releaseFrozen: false, node: { id: 'node03', host: '10.0.0.15',    status: 'degraded' } },
    { name: 'prometheus',   image: 'prom/prometheus:latest',    status: 'stopped',  cpu: 0,  mem: 0,    uptime: 'stopped',restarts: 0,  env: 'stage', incident: false, releaseFrozen: false, node: { id: 'node01', host: '192.168.1.10', status: 'online'   } },
];

const MOCK_NODES = [
    { id: 'node01', host: '192.168.1.10', status: 'online'   as const },
    { id: 'node02', host: '192.168.1.42', status: 'online'   as const },
    { id: 'node03', host: '10.0.0.15',    status: 'degraded' as const },
];

export default function SearchPage() {
    const [query, setQuery] = useState('');
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(function autoFocus() {
        inputRef.current?.focus();
    }, []);

    function handleQueryChange(e: React.ChangeEvent<HTMLInputElement>) {
        setQuery(e.target.value);
    }

    const { matchedServices, matchedNodes } = useMemo(function computeResults() {
        const q = query.trim().toLowerCase();
        if (!q) return { matchedServices: [], matchedNodes: [] };

        const matchedServices = MOCK_SERVICES.filter(s =>
            s.name.toLowerCase().includes(q) ||
            s.image.toLowerCase().includes(q) ||
            s.node.id.toLowerCase().includes(q)
        );

        const matchedNodes = MOCK_NODES.filter(n =>
            n.id.toLowerCase().includes(q) ||
            n.host.toLowerCase().includes(q)
        );

        return { matchedServices, matchedNodes };
    }, [query]);

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
