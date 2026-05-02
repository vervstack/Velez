# T21 — NodeHealthList Widget

**Files to create:**

- `src/widgets/controlplane/NodeHealthList.tsx`
- `src/widgets/controlplane/NodeHealthList.module.css`

## What it looks like

A titled section with a vertical list of `NodeCard` components.

```
NODE HEALTH

[NodeCard — node01]
[NodeCard — node02]
[NodeCard — node03 (degraded)]
```

## Props interface

```ts
import { type NodeCardData } from '@/components/node/NodeCard';

interface NodeHealthListProps {
    nodes: NodeCardData[];
    onShell?: (nodeId: string) => void;
    onDrain?: (nodeId: string) => void;
}
```

## Component

```tsx
// src/widgets/controlplane/NodeHealthList.tsx
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
```

## CSS

```css
/* src/widgets/controlplane/NodeHealthList.module.css */
.NodeHealthListContainer { }

.header {
    margin-bottom: 0.875rem;
}

.list {
    display: flex;
    flex-direction: column;
    gap: 0.625rem;
}
```
