import cls from '@/components/base/StatCard.module.css';

interface StatCardProps {
    value: string | number;
    label: string;
    color?: string;
}

export default function StatCard({ value, label, color = 'var(--cyan)' }: StatCardProps) {
    return (
        <div className={cls.StatCardContainer}>
            <div className={cls.value} style={{ color }}>
                {value}
            </div>
            <div className={cls.label}>{label}</div>
        </div>
    );
}
