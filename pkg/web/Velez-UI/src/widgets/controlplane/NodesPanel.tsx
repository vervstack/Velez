import cls from '@/widgets/controlplane/NodesPanel.module.css';

import NodeHealthList from '@/widgets/controlplane/NodeHealthList';
import NodeTopologyGraph from '@/widgets/controlplane/NodeTopologyGraph';

import {NodeBaseInfo} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins';

const GRAPH_THRESHOLD = 3;

interface NodesPanelProps {
    nodes: NodeBaseInfo[];
    plugins: VervPlugin[];
    onShell?: () => void;
    onDrain?: () => void;
}

export default function NodesPanel({nodes, plugins, onShell, onDrain}: NodesPanelProps) {
    return (
        <div className={cls.NodesPanelContainer}>
            {nodes.length > GRAPH_THRESHOLD
                ? <NodeTopologyGraph nodes={nodes} plugins={plugins}/>
                : <NodeHealthList nodes={nodes} plugins={plugins} onShell={onShell} onDrain={onDrain}/>
            }
        </div>
    );
}
