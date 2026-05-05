import {useState} from 'react';
import cls from '@/widgets/controlplane/PluginMatrix.module.css';
import SectionLabel from '@/components/base/SectionLabel';
import StatusDot from '@/components/base/StatusDot';
import IconButton from '@/components/base/IconButton';
import PluginManageDialog from '@/components/complex/PluginManageDialog/PluginManageDialog';
import {NodeBaseInfo} from "@/app/api/velez";
import {useNavigate} from 'react-router-dom';
import {Routes} from '@/app/router/Routes';

interface PluginStatus {
    pluginName: string;
    tag: string;
    nodeStatuses: Record<string, 'enabled' | 'disabled'>;
    serviceKey?: string;
}

interface PluginMatrixProps {
    nodes: NodeBaseInfo[];
    plugins: PluginStatus[];
    onEnable?: (pluginName: string, nodeId: string) => void;
    onDisable?: (pluginName: string, nodeId: string) => void;
}

export default function PluginMatrix(props: PluginMatrixProps) {
    const [selectedPlugin, setSelectedPlugin] = useState<string | null>(null);

    function handleManagePlugin(pluginName: string) {
        setSelectedPlugin(pluginName);
    }

    function handleCloseDialog() {
        setSelectedPlugin(null);
    }

    return (
        <div className={cls.PluginMatrixContainer}>
            <div className={cls.Header}>
                <SectionLabel>VervStack Plugins</SectionLabel>
            </div>

            <Table {...props} onManagePlugin={handleManagePlugin}/>

            {selectedPlugin && (
                <PluginManageDialog
                    isOpen={true}
                    pluginName={selectedPlugin}
                    onClose={handleCloseDialog}
                />
            )}
        </div>
    );
}

interface TableProps extends PluginMatrixProps {
    onManagePlugin: (pluginName: string) => void;
}

function Table({nodes, plugins, onManagePlugin}: TableProps) {
    const navigate = useNavigate();
    const colTemplate = `180px repeat(${nodes.length}, 1fr) 80px`;

    function renderNodeHeader(n: NodeBaseInfo) {
        return <span key={n.id} className={cls.HeaderCellCenter}>Status</span>;
    }

    function handlePluginNameClick(serviceKey?: string) {
        if (serviceKey) {
            navigate(`${Routes.Service}/${serviceKey}`);
        }
    }

    function renderPlugin(plugin: PluginStatus, pi: number) {
        const isLast = pi === plugins.length - 1;
        const isClickable = !!plugin.serviceKey;

        function renderCell(n: NodeBaseInfo) {
            const status = plugin.nodeStatuses[n.id || ''] ?? 'disabled';

            return (
                <div key={n.id} className={cls.statusCell}>
                    <StatusDot status={status as 'enabled' | 'disabled'}/>
                </div>
            );
        }

        function handleManageClick() {
            onManagePlugin(plugin.pluginName);
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
                <div
                    className={`${cls.PluginCell} ${isClickable ? cls.PluginCellClickable : ''}`}
                    onClick={() => handlePluginNameClick(plugin.serviceKey)}
                >
                    <span className={cls.pluginName}>{plugin.pluginName}</span>
                    <span className={cls.pluginTag}>{plugin.tag}</span>
                </div>

                {nodes.map(renderCell)}

                <div className={cls.manageCell}>
                    <IconButton
                        label="Manage"
                        onClick={handleManageClick}
                        title="Manage plugin"
                    />
                </div>
            </div>
        );
    }

    return (
        <div className={cls.TableContainer}>
            <div className={cls.TableHeader}
                 style={{gridTemplateColumns: colTemplate}}>
                <span className={cls.HeaderCell}>Plugin</span>
                {nodes.map(renderNodeHeader)}
                <span className={cls.HeaderCell}></span>
            </div>

            {plugins.map(renderPlugin)}
        </div>
    )
}
