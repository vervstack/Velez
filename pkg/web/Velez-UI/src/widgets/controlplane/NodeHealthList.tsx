import {NodeBaseInfo} from "@/app/api/velez";

import cls from '@/widgets/controlplane/NodeHealthList.module.css';

import NodeCard from '@/components/node/NodeCard';
import SectionLabel from '@/components/base/SectionLabel';

interface NodeHealthListProps {
    nodes: NodeBaseInfo[];
    onShell?: () => void;
    onDrain?: () => void;
}

export default function NodeHealthList({nodes, onShell, onDrain}: NodeHealthListProps) {
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
    onShell?: () => void;
    onDrain?: () => void;
}

function Node({
                  node,
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
            onShell={handleShell}
            onDrain={handleDrain}
        />
    );
}
