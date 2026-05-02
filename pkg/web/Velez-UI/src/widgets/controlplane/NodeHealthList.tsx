import cls from '@/widgets/controlplane/NodeHealthList.module.css';
import NodeCard, { type NodeCardData } from '@/components/node/NodeCard';
import SectionLabel from '@/components/base/SectionLabel';

interface NodeHealthListProps {
    nodes: NodeCardData[];
    onShell?: (nodeId: string) => void;
    onDrain?: (nodeId: string) => void;
}

export default function NodeHealthList({ nodes, onShell, onDrain }: NodeHealthListProps) {
    return (
        <div className={cls.NodeHealthListContainer}>
            <div className={cls.header}>
                <SectionLabel>Node Health</SectionLabel>
            </div>
            <div className={cls.list}>
                {nodes.map(function renderNode(node) {
                    function handleShell() { onShell?.(node.id); }
                    function handleDrain() { onDrain?.(node.id); }
                    return (
                        <NodeCard
                            key={node.id}
                            node={node}
                            onShell={handleShell}
                            onDrain={handleDrain}
                        />
                    );
                })}
            </div>
        </div>
    );
}
