import { Smerd, SmerdStatus } from '@/app/api/velez';
import { ServiceCardData } from '@/components/service/ServiceCard';

export interface AppData {
    name: string;
    image: string;
    status: 'running' | 'degraded' | 'stopped';
    env: string;
    incident: boolean;
    releaseFrozen: boolean;
    node: { id: string; host: string; status: 'online' | 'degraded' | 'offline' };
    deployments: number;
    lastDeployed: string;
    configSource: string;
    version: string;
}

export const LOCAL_NODE = { id: 'local', host: 'localhost', status: 'online' as const };

export function mapSmerdToServiceCard(s: Smerd): ServiceCardData {
    return {
        name: s.name ?? '',
        image: s.imageName ?? '',
        status: mapSmerdStatus(s.status),
        cpu: 0,
        mem: 0,
        uptime: '-',
        restarts: 0,
        env: s.labels?.['env'] ?? 'prod',
        incident: false,
        releaseFrozen: false,
        node: LOCAL_NODE,
    };
}

export function mapSmerdToAppData(s: Smerd): AppData {
    return {
        name: s.name ?? '',
        image: s.imageName ?? '',
        status: mapSmerdStatus(s.status),
        env: s.labels?.['env'] ?? 'prod',
        incident: false,
        releaseFrozen: false,
        node: LOCAL_NODE,
        deployments: 0,
        lastDeployed: '-',
        configSource: 'none',
        version: s.imageName?.split(':')[1] ?? 'latest',
    };
}

function mapSmerdStatus(s?: SmerdStatus): 'running' | 'degraded' | 'stopped' {
    switch (s) {
        case SmerdStatus.running:
            return 'running';
        case SmerdStatus.restarting:
            return 'degraded';
        case SmerdStatus.dead:
        case SmerdStatus.exited:
        case SmerdStatus.paused:
        case SmerdStatus.removing:
            return 'stopped';
        default:
            return 'stopped';
    }
}
