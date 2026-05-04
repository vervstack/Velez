import cls from '@/widgets/controlplane/PluginMatrix.module.css';
import Badge from '@/components/base/Badge';
import SectionLabel from '@/components/base/SectionLabel';
import {NodeBaseInfo} from "@/app/api/velez";

interface PluginStatus {
    pluginName: string;
    tag: string;
    nodeStatuses: Record<string, 'enabled' | 'disabled'>;
}

interface PluginMatrixProps {
    nodes: NodeBaseInfo[];
    plugins: PluginStatus[];
}

export default function PluginMatrix(props: PluginMatrixProps) {
    return (
        <div className={cls.PluginMatrixContainer}>
            <div className={cls.Header}>
                <SectionLabel>VervStack Plugins</SectionLabel>
            </div>

            <Table {...props}/>
        </div>
    );
}

function Table({nodes, plugins}: PluginMatrixProps) {
    const colTemplate = `180px repeat(${nodes.length}, 1fr)`;

    function renderNodeHeader(n: NodeBaseInfo) {
        return <span key={n.id} className={cls.HeaderCellCenter}>{n.id}</span>;
    }

    function renderPlugin(plugin: PluginStatus, pi: number) {
        const isLast = pi === plugins.length - 1;

        function renderCell(n: NodeBaseInfo) {
            const status = plugin.nodeStatuses[n.id || ''] ?? 'disabled';
            const color = status === 'enabled' ? 'var(--green)' : 'var(--fg-dim)';
            return (
                <div key={n.id} className={cls.statusCell}>
                    <Badge label={status} color={color}/>
                </div>
            );
        }

        return (
            <div
                key={plugin.pluginName}
                className={cls.TableRow}
                style={{
                    gridTemplateColumns: colTemplate,
                    borderBottom: isLast ? 'none' : undefined,
                }}
            >
                <div className={cls.PluginCell}>
                    <span className={cls.pluginName}>{plugin.pluginName}</span>
                    <span className={cls.pluginTag}>{plugin.tag}</span>
                </div>

                {nodes.map(renderCell)}
            </div>
        );
    }

    return (
        <div className={cls.TableContainer}>
            <div className={cls.TableHeader}
                 style={{gridTemplateColumns: colTemplate}}>
                <span className={cls.HeaderCell}>Plugin</span>
                {nodes.map(renderNodeHeader)}
            </div>

            {plugins.map(renderPlugin)}
        </div>
    )
}
