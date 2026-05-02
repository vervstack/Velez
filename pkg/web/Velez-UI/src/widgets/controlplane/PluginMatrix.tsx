import cls from '@/widgets/controlplane/PluginMatrix.module.css';
import Badge from '@/components/base/Badge';
import SectionLabel from '@/components/base/SectionLabel';

interface PluginStatus {
    pluginName: string;
    tag: string;
    nodeStatuses: Record<string, 'enabled' | 'disabled'>;
}

interface PluginMatrixProps {
    nodes: Array<{ id: string }>;
    plugins: PluginStatus[];
}

export default function PluginMatrix({ nodes, plugins }: PluginMatrixProps) {
    const colTemplate = `180px repeat(${nodes.length}, 1fr)`;

    return (
        <div className={cls.PluginMatrixContainer}>
            <div className={cls.header}>
                <SectionLabel>VervStack Plugins</SectionLabel>
            </div>
            <div className={cls.table}>
                {/* Table header */}
                <div className={cls.tableHeader} style={{ gridTemplateColumns: colTemplate }}>
                    <span className={cls.headerCell}>Plugin</span>
                    {nodes.map(function renderNodeHeader(n) {
                        return <span key={n.id} className={cls.headerCellCenter}>{n.id}</span>;
                    })}
                </div>

                {/* Rows */}
                {plugins.map(function renderPlugin(plugin, pi) {
                    const isLast = pi === plugins.length - 1;
                    return (
                        <div
                            key={plugin.pluginName}
                            className={cls.tableRow}
                            style={{
                                gridTemplateColumns: colTemplate,
                                borderBottom: isLast ? 'none' : undefined,
                            }}
                        >
                            <div className={cls.pluginCell}>
                                <span className={cls.pluginName}>{plugin.pluginName}</span>
                                <span className={cls.pluginTag}>{plugin.tag}</span>
                            </div>
                            {nodes.map(function renderCell(n) {
                                const status = plugin.nodeStatuses[n.id] ?? 'disabled';
                                const color = status === 'enabled' ? 'var(--green)' : 'var(--fg-dim)';
                                return (
                                    <div key={n.id} className={cls.statusCell}>
                                        <Badge label={status} color={color} />
                                    </div>
                                );
                            })}
                        </div>
                    );
                })}
            </div>
        </div>
    );
}
