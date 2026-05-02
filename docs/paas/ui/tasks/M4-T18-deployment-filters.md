# T18 — DeploymentFilters Widget (toolbar)

**Files to create:**

- `src/widgets/deployments/DeploymentFilters.tsx`
- `src/widgets/deployments/DeploymentFilters.module.css`

## What it looks like

A toolbar row with search input, status filter chips, env filter chips, a clear-all button, a view-mode toggle, and a
count label.

```
[⌕ search services…]  |  [● Running] [● Degraded] [● Stopped]  |  [prod] [stage]  [✕ clear]          [⬛⬛ Kanban] [≡ List]  9 services
```

## Props interface

```ts
type ViewMode = 'kanban' | 'list';

interface DeploymentFiltersProps {
    search: string;
    onSearchChange: (v: string) => void;
    statusFilters: Set<string>;
    onToggleStatus: (id: string) => void;
    envFilters: Set<string>;
    onToggleEnv: (id: string) => void;
    onClearAll: () => void;
    viewMode: ViewMode;
    onViewModeChange: (v: ViewMode) => void;
    totalCount: number;
}
```

## Component

```tsx
// src/widgets/deployments/DeploymentFilters.tsx
import cls from '@/widgets/deployments/DeploymentFilters.module.css';
import cn from 'classnames';

type ViewMode = 'kanban' | 'list';

interface DeploymentFiltersProps {
    search: string;
    onSearchChange: (v: string) => void;
    statusFilters: Set<string>;
    onToggleStatus: (id: string) => void;
    envFilters: Set<string>;
    onToggleEnv: (id: string) => void;
    onClearAll: () => void;
    viewMode: ViewMode;
    onViewModeChange: (v: ViewMode) => void;
    totalCount: number;
}

const STATUS_OPTIONS = [
    { id: 'running',  label: 'Running',  color: 'var(--green)'   },
    { id: 'degraded', label: 'Degraded', color: 'var(--amber)'   },
    { id: 'stopped',  label: 'Stopped',  color: 'var(--fg-faint)'},
];

const ENV_OPTIONS = [
    { id: 'prod',  label: 'prod',  color: 'var(--cyan)'   },
    { id: 'stage', label: 'stage', color: 'var(--fg-dim)' },
];

export default function DeploymentFilters({
    search,
    onSearchChange,
    statusFilters,
    onToggleStatus,
    envFilters,
    onToggleEnv,
    onClearAll,
    viewMode,
    onViewModeChange,
    totalCount,
}: DeploymentFiltersProps) {
    const hasActiveFilters = statusFilters.size > 0 || envFilters.size > 0 || search !== '';

    function handleSearchChange(e: React.ChangeEvent<HTMLInputElement>) {
        onSearchChange(e.target.value);
    }

    return (
        <div className={cls.DeploymentFiltersContainer}>
            {/* Search */}
            <div className={cls.searchWrapper}>
                <span className={cls.searchIcon}>⌕</span>
                <input
                    className={cls.searchInput}
                    value={search}
                    onChange={handleSearchChange}
                    placeholder="search services…"
                />
            </div>

            <div className={cls.sep} />

            {/* Status chips */}
            <div className={cls.chipGroup}>
                {STATUS_OPTIONS.map(function renderStatus(opt) {
                    const active = statusFilters.has(opt.id);
                    function handleClick() { onToggleStatus(opt.id); }
                    return (
                        <button
                            key={opt.id}
                            className={cn(cls.filterChip, { [cls.filterChipActive]: active })}
                            style={active ? {
                                borderColor: opt.color + '66',
                                background: opt.color + '18',
                                color: opt.color,
                            } : undefined}
                            onClick={handleClick}
                        >
                            <span
                                className={cls.chipDot}
                                style={{ background: active ? opt.color : 'var(--fg-faint)' }}
                            />
                            {opt.label}
                        </button>
                    );
                })}
            </div>

            <div className={cls.sep} />

            {/* Env chips */}
            <div className={cls.chipGroup}>
                {ENV_OPTIONS.map(function renderEnv(opt) {
                    const active = envFilters.has(opt.id);
                    function handleClick() { onToggleEnv(opt.id); }
                    return (
                        <button
                            key={opt.id}
                            className={cn(cls.filterChip, { [cls.filterChipActive]: active })}
                            style={active ? {
                                borderColor: opt.color + '66',
                                background: opt.color + '18',
                                color: opt.color,
                            } : undefined}
                            onClick={handleClick}
                        >
                            {opt.label}
                        </button>
                    );
                })}
            </div>

            {/* Clear */}
            {hasActiveFilters && (
                <button className={cls.clearBtn} onClick={onClearAll}>✕ clear</button>
            )}

            {/* View mode */}
            <div className={cls.viewToggle}>
                <button
                    className={cn(cls.viewBtn, { [cls.viewBtnActive]: viewMode === 'kanban' })}
                    onClick={() => onViewModeChange('kanban')}
                >
                    kanban
                </button>
                <button
                    className={cn(cls.viewBtn, { [cls.viewBtnActive]: viewMode === 'list' })}
                    onClick={() => onViewModeChange('list')}
                >
                    list
                </button>
            </div>

            <span className={cls.count}>{totalCount} services</span>
        </div>
    );
}
```

## CSS

```css
/* src/widgets/deployments/DeploymentFilters.module.css */
.DeploymentFiltersContainer {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.75rem 1.5rem 0;
    flex-shrink: 0;
    flex-wrap: wrap;
}

/* Search */
.searchWrapper {
    position: relative;
    flex-shrink: 0;
}

.searchIcon {
    position: absolute;
    left: 0.5625rem;
    top: 50%;
    transform: translateY(-50%);
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--fg-faint);
    pointer-events: none;
}

.searchInput {
    background: var(--bg3);
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    padding: 0.375rem 0.625rem 0.375rem 1.625rem;
    font-family: var(--font-mono);
    font-size: 0.71875rem;
    color: var(--fg);
    width: 12.5rem;
    outline: none;
    transition: border-color 0.15s;

    &::placeholder { color: var(--fg-faint); }
    &:focus { border-color: var(--border-h); }
}

/* Separator */
.sep {
    width: 1px;
    height: 1.25rem;
    background: var(--border);
    flex-shrink: 0;
}

/* Chip groups */
.chipGroup {
    display: flex;
    gap: 0.3125rem;
    align-items: center;
}

.filterChip {
    display: inline-flex;
    align-items: center;
    gap: 0.3125rem;
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    padding: 0.25rem 0.625rem;
    border-radius: 99px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--fg-dim);
    cursor: pointer;
    transition: all 0.12s;

    &:hover:not(.filterChipActive) {
        border-color: var(--border-m);
        color: var(--fg-mid);
    }
}

.filterChipActive { }  /* colors applied inline */

.chipDot {
    width: 0.3125rem;
    height: 0.3125rem;
    border-radius: 50%;
    display: block;
    transition: background 0.12s;
}

/* Clear */
.clearBtn {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    padding: 0.1875rem 0.5625rem;
    border-radius: 99px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--fg-dim);
    cursor: pointer;
    transition: all 0.12s;

    &:hover {
        border-color: var(--border-m);
        color: var(--fg);
    }
}

/* View mode toggle */
.viewToggle {
    display: flex;
    gap: 0.0625rem;
    background: var(--bg3);
    border-radius: 0.5rem;
    padding: 0.125rem;
    border: 1px solid var(--border);
    margin-left: auto;
}

.viewBtn {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    padding: 0.3125rem 0.75rem;
    border-radius: 0.375rem;
    border: none;
    cursor: pointer;
    background: transparent;
    color: var(--fg-dim);
    transition: all 0.12s;
}

.viewBtnActive {
    background: var(--bg2);
    color: var(--fg);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

/* Count */
.count {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
    flex-shrink: 0;
}
```

## Notes

- Inline `style` on `.filterChip` is acceptable here because the chip colors are runtime-parametric (each chip has a
  different accent color from the options array).
- The `margin-left: auto` on `.viewToggle` pushes the toggle and count to the right end.
