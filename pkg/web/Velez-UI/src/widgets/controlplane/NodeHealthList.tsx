import {NodeBaseInfo} from "@/app/api/velez";
import {VervPlugin} from "@/model/services/VervPlugins";
import cls from '@/widgets/controlplane/NodeHealthList.module.css';
import NodeCard from '@/components/node/NodeCard';
import SectionLabel from '@/components/base/SectionLabel';

interface NodeHealthListProps {
    nodes: NodeBaseInfo[];
    plugins: VervPlugin[];
    onShell?: () => void;
    onDrain?: () => void;
}

export default function NodeHealthList({nodes, plugins, onShell, onDrain}: NodeHealthListProps) {
    return (
        <div className={cls.NodeHealthListContainer}>
            <div className={cls.Header}>
                <SectionLabel>Node Health</SectionLabel>
            </div>

            <div className={cls.list}>
                {nodes
                    .map(n =>
                        <div
                            key={n.id}
                        >
                            <Node
                                node={n}
                                plugins={plugins.filter(p => n.id !== undefined && p.nodeIds.includes(n.id))}
                                onShell={onShell}
                                onDrain={onDrain}/>
                        </div>
                    )}
            </div>
        </div>
    );
}

interface NodeProps {
    node: NodeBaseInfo;
    plugins: VervPlugin[];
    onShell?: () => void;
    onDrain?: () => void;
}

function Node({
                  node,
                  plugins,
                  onShell,
                  onDrain,
              }: NodeProps) {
    function handleShell() {
        onShell?.();
    }

    function handleDrain() {
        onDrain?.();
    }

    return (
        <NodeCard
            key={node.id}
            node={node}
            plugins={plugins}
            onShell={handleShell}
            onDrain={handleDrain}
        />
    );
}
