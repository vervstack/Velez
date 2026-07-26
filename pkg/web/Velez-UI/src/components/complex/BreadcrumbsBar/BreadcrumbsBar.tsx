import cls from '@/components/complex/BreadcrumbsBar/BreadcrumbsBar.module.css';

export interface Crumb {
    label: string;
    onClick?: () => void;
}

interface BreadcrumbsBarProps {
    crumbs: Crumb[];
}

export default function BreadcrumbsBar({crumbs}: BreadcrumbsBarProps) {
    return (
        <div className={cls.BreadcrumbsBarContainer}>
            {crumbs.map(function renderCrumb(crumb, idx) {
                const isLast = idx === crumbs.length - 1;
                return (
                    <span key={crumb.label}>
                        {isLast ? (
                            <span className={cls.BreadcrumbCurrent}>{crumb.label}</span>
                        ) : (
                            <span className={cls.BreadcrumbLink} onClick={crumb.onClick}>{crumb.label}</span>
                        )}
                        {!isLast && <span className={cls.BreadcrumbSep}>/</span>}
                    </span>
                );
            })}
        </div>
    );
}
