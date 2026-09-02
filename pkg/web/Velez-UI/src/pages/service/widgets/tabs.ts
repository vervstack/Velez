export type ServiceTab = 'overview' | 'metrics' | 'instances' | 'history' | 'access';

export interface TabInfo {
    id: ServiceTab;
    label: string;
}

export const TABS: TabInfo[] = [
    {id: 'overview', label: 'Overview'},
    {id: 'metrics', label: 'Metrics'},
    {id: 'instances', label: 'Instances'},
    {id: 'history', label: 'History'},
    {id: 'access', label: 'Access'},
];
