import cls from '@/components/base/Badge.module.css';

interface BadgeProps {
    label: string;
    color?: string;
    dim?: string;
}

export default function Badge({ label, color, dim }: BadgeProps) {
    return (
        <span
            className={cls.BadgeContainer}
            style={{
                color: color ?? 'var(--fg-dim)',
                borderColor: color ? color + '44' : 'var(--border)',
                background: dim ?? 'transparent',
            }}
        >
            {label}
        </span>
    );
}
