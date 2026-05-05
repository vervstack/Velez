import cls from '@/widgets/controlplane/PluginMatrix.module.css';
import SectionLabel from '@/components/base/SectionLabel';
import IconButton from '@/components/base/IconButton';
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
    return (
        <div className={cls.PluginMatrixContainer}>
            <div className={cls.Header}>
                <SectionLabel>VervStack Plugins</SectionLabel>
            </div>

            <Table {...props}/>
        </div>
    );
}

function Table({nodes, plugins, onEnable, onDisable}: PluginMatrixProps) {
    const navigate = useNavigate();
    const colTemplate = `180px repeat(${nodes.length}, 1fr)`;

    function renderNodeHeader(n: NodeBaseInfo) {
        return <span key={n.id} className={cls.HeaderCellCenter}>{n.id}</span>;
    }

    function handlePluginNameClick(serviceKey?: string) {
        if (serviceKey) {
            navigate(`${Routes.Service}/${serviceKey}`);
        }
    }

    function renderPlugin(plugin: PluginStatus, pi: number) {
        const isLast = pi === plugins.length - 1;
        const isClickable = !!plugin.serviceKey;

        function handleEnableClick(nodeId: string | undefined) {
            if (nodeId && onEnable) {
                onEnable(plugin.pluginName, nodeId);
            }
        }

        function handleDisableClick(nodeId: string | undefined) {
            if (nodeId && onDisable) {
                onDisable(plugin.pluginName, nodeId);
            }
        }

        function renderCell(n: NodeBaseInfo) {
            const status = plugin.nodeStatuses[n.id || ''] ?? 'disabled';
            const isEnabled = status === 'enabled';

            function handleButtonClick() {
                if (isEnabled) {
                    handleDisableClick(n.id);
                } else {
                    handleEnableClick(n.id);
                }
            }

            const buttonLabel = isEnabled ? '✓' : '✕';
            const buttonTitle = isEnabled ? 'Disable' : 'Enable';

            return (
                <div key={n.id} className={cls.statusCell}>
                    <IconButton
                        label={buttonLabel}
                        title={buttonTitle}
                        onClick={handleButtonClick}
                        danger={isEnabled}
                    />
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
                <div
                    className={`${cls.PluginCell} ${isClickable ? cls.PluginCellClickable : ''}`}
                    onClick={() => handlePluginNameClick(plugin.serviceKey)}
                >
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
