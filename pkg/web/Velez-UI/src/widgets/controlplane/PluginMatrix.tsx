import {useNavigate} from 'react-router-dom';

import {NodeBaseInfo, VervPluginState, VervPluginType} from "@/app/api/velez";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";

import cls from '@/widgets/controlplane/PluginMatrix.module.css';

import SectionLabel from '@/components/base/SectionLabel';
import StatusDot from '@/components/base/StatusDot';
import IconButton from '@/components/base/IconButton';
import PluginManageDialog from '@/dialogs/PluginManageDialog/PluginManageDialog';
import {Routes} from '@/app/router/Routes';
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";


interface PluginMatrixProps {
    nodes: NodeBaseInfo[];
    plugins: VervPlugin[];
    onEnable?: (pluginType: VervPluginType, payload?: never) => Promise<void>;
    onDisable?: (pluginName: string, nodeId: string) => void;
}

export default function PluginMatrix(props: PluginMatrixProps) {
    const {OpenDialog} = useDialog()


    function handleManagePlugin(plugin: VervPlugin) {
        OpenDialog(<PluginManageDialog pluginType={plugin.type}/>)
    }


    return (
        <div className={cls.PluginMatrixContainer}>
            <div className={cls.Header}>
                <SectionLabel>VervStack Plugins</SectionLabel>
            </div>

            <Table {...props} onManagePlugin={handleManagePlugin}/>
        </div>
    );
}

interface TableProps extends PluginMatrixProps {
    onManagePlugin: (plugin: VervPlugin) => void;
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

    function renderPlugin(plugin: VervPlugin, pi: number) {
        const isLast = pi === plugins.length - 1;
        const isDisabled = plugin.state === VervPluginState.disabled;

        function renderCell(n: NodeBaseInfo) {
            return (
                <div
                    key={n.id}
                    className={cls.statusCell}>
                    <StatusDot
                        status={mapVervPluginStateToPluginStatus(plugin.state)}/>
                </div>
            );
        }

        function handleManageClick() {
            onManagePlugin(plugin);
        }

        function openServicePage(serviceName: string) {
            navigate(`${Routes.Service}/${serviceName}`);
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
                <div
                    className={`${cls.PluginCell} ${isDisabled ? '' : cls.PluginCellClickable}`}
                    onClick={isDisabled ? undefined : () => handlePluginNameClick(plugin.type)}
                >
                    <span className={cls.pluginName}>{plugin.title}</span>
                </div>

                {nodes.map(renderCell)}

                <div className={cls.manageCell}>
                    {
                        plugin.state == VervPluginState.running ?
                            <IconButton
                                label={"Open service"}
                                onClick={() => openServicePage(plugin.serviceName)}
                            />
                            :
                            <IconButton
                                label="Manage"
                                onClick={handleManageClick}
                                title="Manage plugin"
                            />
                    }
                </div>
            </div>)
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


function mapVervPluginStateToPluginStatus(state?: VervPluginState): 'running' | 'degraded' | 'stopped' | 'online' | 'offline' | 'enabled' | 'disabled' {
    switch (state) {
        case VervPluginState.running:
            return 'enabled';
        case VervPluginState.disabled:
            return 'disabled';
        case VervPluginState.warning:
        case VervPluginState.dead:
            return 'degraded';
        default:
            return 'disabled';
    }
}
