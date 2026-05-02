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
