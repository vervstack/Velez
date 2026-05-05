import {useState} from 'react';
import {useNavigate} from 'react-router-dom';

import {NodeBaseInfo, VervService, VervServiceState, VervServiceType} from "@/app/api/velez";

import cls from '@/widgets/controlplane/PluginMatrix.module.css';

import SectionLabel from '@/components/base/SectionLabel';
import StatusDot from '@/components/base/StatusDot';
import IconButton from '@/components/base/IconButton';
import PluginManageDialog from '@/components/complex/PluginManageDialog/PluginManageDialog';
import {Routes} from '@/app/router/Routes';


interface PluginMatrixProps {
    nodes: NodeBaseInfo[];
    plugins: VervService[];
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
    onManagePlugin: (pluginName: VervServiceType) => void;
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

    function renderPlugin(plugin: VervService, pi: number) {
        const isLast = pi === plugins.length - 1;

        function renderCell(n: NodeBaseInfo) {
            return (
                <div
                    key={n.id}
                    className={cls.statusCell}>
                    <StatusDot
                        status={mapVervServiceStateToPluginStatus(plugin.state)}/>
                </div>
            );
        }

        function handleManageClick() {
            onManagePlugin(plugin.type || VervServiceType.unknown_service_type);
        }

        return (
            <div
                key={plugin.type}
                className={cls.TableRow}
                style={{
                    gridTemplateColumns: colTemplate,
                    borderBottom: isLast ? 'none' : undefined,
                }}
            >
                {/*TODO add logic on open dialog when clicked*/}
                <div
                    className={`${cls.PluginCell} ${cls.PluginCellClickable}`}
                    onClick={() => handlePluginNameClick(plugin.type)}
                >
                    <span className={cls.pluginName}>{plugin.type?.toString()}</span>
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


function mapVervServiceStateToPluginStatus(state?: VervServiceState): 'running' | 'degraded' | 'stopped' | 'online' | 'offline' | 'enabled' | 'disabled' {
    switch (state) {
        case VervServiceState.running:
            return 'enabled';
        case VervServiceState.disabled:
            return 'disabled';
        case VervServiceState.warning:
        case VervServiceState.dead:
            return 'degraded';
        default:
            return 'disabled';
    }
}
