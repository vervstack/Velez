# T22 — PluginMatrix Widget

**Files to create:**

- `src/widgets/controlplane/PluginMatrix.tsx`
- `src/widgets/controlplane/PluginMatrix.module.css`

## What it looks like

A table with plugins as rows and nodes as columns. Each cell shows an "enabled"/"disabled" badge.

```
VERVSTACK PLUGINS

┌──────────────────────────────────────┐
│ Plugin       │ node01 │ node02 │ node03 │
│─────────────────────────────────────│
│ Matreshka    │ [enabled] │ [enabled] │ [enabled] │
│ Makosh       │ [enabled] │ [enabled] │ [disabled] │
│ Svarog       │ [enabled] │ [enabled] │ [enabled] │
└──────────────────────────────────────┘
```

## Data types

```ts
interface PluginStatus {
    pluginName: string;
    tag: string;       // e.g. "config", "gRPC", "secrets"
    nodeStatuses: Record<string, 'enabled' | 'disabled'>; // nodeId → status
}
```

## Props interface

```ts
interface PluginMatrixProps {
    nodes: Array<{ id: string }>;
    plugins: PluginStatus[];
}
```

## Component

```tsx
// src/widgets/controlplane/PluginMatrix.tsx
import cls from '@/widgets/controlplane/PluginMatrix.module.css';
import Badge from '@/components/base/Badge';
import SectionLabel from '@/components/base/SectionLabel';

interface PluginStatus {
    pluginName: string;
    tag: string;
    nodeStatuses: Record<string, 'enabled' | 'disabled'>;
}

interface PluginMatrixProps {
    nodes: Array<{ id: string }>;
    plugins: PluginStatus[];
}

export default function PluginMatrix({ nodes, plugins }: PluginMatrixProps) {
    const colTemplate = `180px repeat(${nodes.length}, 1fr)`;

    return (
        <div className={cls.PluginMatrixContainer}>
            <div className={cls.header}>
                <SectionLabel>VervStack Plugins</SectionLabel>
            </div>
            <div className={cls.table}>
                {/* Table header */}
                <div className={cls.tableHeader} style={{ gridTemplateColumns: colTemplate }}>
                    <span className={cls.headerCell}>Plugin</span>
                    {nodes.map(function renderNodeHeader(n) {
                        return <span key={n.id} className={cls.headerCellCenter}>{n.id}</span>;
                    })}
                </div>

                {/* Rows */}
                {plugins.map(function renderPlugin(plugin, pi) {
                    const isLast = pi === plugins.length - 1;
                    return (
                        <div
                            key={plugin.pluginName}
                            className={cls.tableRow}
                            style={{
                                gridTemplateColumns: colTemplate,
                                borderBottom: isLast ? 'none' : undefined,
                            }}
                        >
                            <div className={cls.pluginCell}>
                                <span className={cls.pluginName}>{plugin.pluginName}</span>
                                <span className={cls.pluginTag}>{plugin.tag}</span>
                            </div>
                            {nodes.map(function renderCell(n) {
                                const status = plugin.nodeStatuses[n.id] ?? 'disabled';
                                const color = status === 'enabled' ? 'var(--green)' : 'var(--fg-dim)';
                                return (
                                    <div key={n.id} className={cls.statusCell}>
                                        <Badge label={status} color={color} />
                                    </div>
                                );
                            })}
                        </div>
                    );
                })}
            </div>
        </div>
    );
}
```

## CSS

```css
/* src/widgets/controlplane/PluginMatrix.module.css */
.PluginMatrixContainer { }

.header {
    margin-bottom: 0.875rem;
}

.table {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
}

.tableHeader {
    display: grid;
    padding: 0.75rem 1.25rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg3);
}

.headerCell {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
}

.headerCellCenter {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
    text-align: center;
}

.tableRow {
    display: grid;
    padding: 0.875rem 1.25rem;
    border-bottom: 1px solid var(--border);
    align-items: center;
}

.pluginCell {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.pluginName {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--fg);
    font-weight: 500;
}

.pluginTag {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
}

.statusCell {
    display: flex;
    justify-content: center;
}
```

## Notes

- `gridTemplateColumns` is set inline because it's computed from the number of nodes, which is dynamic.
