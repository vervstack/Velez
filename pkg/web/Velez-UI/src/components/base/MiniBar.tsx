import cls from '@/components/base/MiniBar.module.css';

interface MiniBarProps {
    val: number;
    max?: number;
    color?: string;
}

export default function MiniBar({ val, max = 100, color = 'var(--cyan)' }: MiniBarProps) {
    const pct = Math.min(100, (val / max) * 100);
    const fillColor = pct > 80 ? 'var(--red)' : pct > 60 ? 'var(--amber)' : color;

    return (
        <div className={cls.MiniBarContainer}>
            <div
                className={cls.fill}
                style={{ width: `${pct}%`, background: fillColor }}
            />
        </div>
    );
}
