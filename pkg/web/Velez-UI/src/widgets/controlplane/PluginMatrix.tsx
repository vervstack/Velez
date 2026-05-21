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
    return (
        <div className={cls.PluginMatrixContainer}>
            <div className={cls.Header}>
                <SectionLabel>VervStack Plugins</SectionLabel>
            </div>

            <Table {...props}/>
        </div>
    );
}

interface TableProps extends PluginMatrixProps {
}

function Table({nodes, plugins}: TableProps) {

    const colTemplate = `180px repeat(${nodes.length}, 1fr) 80px`;

    return (
        <div className={cls.TableContainer}>
            <div className={cls.TableHeader}
                 style={{gridTemplateColumns: colTemplate}}>

                <span className={cls.HeaderCell}>Plugin</span>

                {nodes.map(NodeHeader)}

                <span className={cls.HeaderCell}></span>
            </div>

            {plugins.map(v => PluginContent(v, colTemplate))}
        </div>
    )
}


function NodeHeader(n: NodeBaseInfo) {
    return (
        <span key={n.id} className={cls.HeaderCellCenter}>
            Status
        </span>)
        ;
}


function PluginContent(
    plugin: VervPlugin,
    colTemplate: string,
) {
    const navigate = useNavigate();
    const {OpenDialog} = useDialog();

    const isDisabled = plugin.state === VervPluginState.disabled;

    function handleManagePlugin(plugin: VervPlugin) {
        OpenDialog(<PluginManageDialog pluginType={plugin.type}/>)
    }

    function handlePluginNameClick(serviceKey?: string) {
        if (serviceKey) {
            navigate(`${Routes.Service}/${serviceKey}`);
        }
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
            }}
        >
            <div
                className={`${cls.PluginCell} ${isDisabled ? '' : cls.PluginCellClickable}`}
                onClick={isDisabled ? undefined : () => handlePluginNameClick(plugin.type)}
            >
                <span className={cls.pluginName}>{plugin.title}</span>
            </div>

            <div className={cls.statusCell}>
                <StatusDot
                    status={mapVervPluginStateToPluginStatus(plugin.state)}/>
            </div>

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
                            onClick={() => handleManagePlugin(plugin)}
                            title="Manage plugin"
                        />
                }
            </div>
        </div>)
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
