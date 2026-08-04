import {useLayoutEffect, useRef} from 'react';
import {useNavigate} from 'react-router-dom';

import {NodeBaseInfo, VervPluginState, VervPluginType} from "@/app/api/velez";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";

import cls from '@/widgets/controlplane/PluginMatrix.module.css';

import SectionLabel from '@/components/base/SectionLabel';
import StatusDot from '@/components/base/StatusDot';
import IconButton from '@/components/base/IconButton';
import PluginManageDialog from '@/dialogs/PluginManageDialog/PluginManageDialog';
import {openStatefullPgDialog} from '@/dialogs/PluginManageDialog/plugins/openStatefullPgDialog.tsx';
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

function getStatusRank(state?: VervPluginState): number {
    switch (state) {
        case VervPluginState.running:
            return 0;
        case VervPluginState.warning:
        case VervPluginState.dead:
            return 1;
        default:
            return 2;
    }
}

function sortPluginsByStatus(plugins: VervPlugin[]): VervPlugin[] {
    return [...plugins].sort((a, b) => getStatusRank(a.state) - getStatusRank(b.state));
}

function Table({nodes, plugins}: TableProps) {

    const colTemplate = `180px repeat(${nodes.length}, 1fr) 80px`;
    const sortedPlugins = sortPluginsByStatus(plugins);
    const orderKey = sortedPlugins.map(v => v.type).join('|');

    const rowNodes = useRef(new Map<string, HTMLDivElement>());
    const prevRects = useRef(new Map<string, DOMRect>());

    function setRowRef(key: string, node: HTMLDivElement | null) {
        if (node) {
            rowNodes.current.set(key, node);
        } else {
            rowNodes.current.delete(key);
        }
    }

    useLayoutEffect(() => {
        const nextRects = new Map<string, DOMRect>();

        rowNodes.current.forEach((node, key) => {
            const newRect = node.getBoundingClientRect();
            nextRects.set(key, newRect);

            const oldRect = prevRects.current.get(key);
            if (!oldRect) {
                return;
            }

            const deltaY = oldRect.top - newRect.top;
            if (!deltaY) {
                return;
            }

            node.style.transition = 'none';
            node.style.transform = `translateY(${deltaY}px)`;

            requestAnimationFrame(() => {
                node.style.transition = 'transform 200ms ease';
                node.style.transform = 'none';
            });
        });

        prevRects.current = nextRects;
    }, [orderKey]);

    return (
        <div className={cls.TableContainer}>
            <div className={cls.TableHeader}
                 style={{gridTemplateColumns: colTemplate}}>

                <span className={cls.HeaderCell}>Plugin</span>

                {nodes.map(NodeHeader)}

                <span className={cls.HeaderCell}></span>
            </div>

            {sortedPlugins.map(v => (
                <div key={v.type} className={cls.RowWrapper} ref={node => setRowRef(v.type, node)}>
                    <PluginContent plugin={v} colTemplate={colTemplate}/>
                </div>
            ))}
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


interface PluginContentProps {
    plugin: VervPlugin;
    colTemplate: string;
}

function PluginContent({plugin, colTemplate}: PluginContentProps) {
    const navigate = useNavigate();
    const {OpenDialog} = useDialog();

    const isDisabled = plugin.state === VervPluginState.disabled;

    function handleManagePlugin(plugin: VervPlugin) {
        if (plugin.type == VervPluginType.statefull_pg) {
            openStatefullPgDialog(plugin);
            return;
        }
        OpenDialog(<PluginManageDialog pluginType={plugin.type}/>)
    }


    function openServicePage(serviceName: string) {
        navigate(`${Routes.Service}/${serviceName}`);
    }

    return (
        <div
            className={cls.TableRow}
            style={{
                gridTemplateColumns: colTemplate,
            }}
        >
            <div
                className={`${cls.PluginCell} ${isDisabled ? '' : cls.PluginCellClickable}`}
                onClick={isDisabled ? undefined : () => openServicePage(plugin.serviceName)}
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
