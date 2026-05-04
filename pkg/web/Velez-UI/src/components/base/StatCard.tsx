import cls from '@/components/base/StatCard.module.css';
import cn from "classnames";

export interface StatCardProps {
    value: string | number;
    label: string;
    level?: Level;
}

export enum Level {
    INFO = 'info',
    WARN = 'warn',
    ERROR = 'error',
    Good = 'good',
}

export default function StatCard(
    {
        value, label,
        level = Level.INFO
    }: StatCardProps) {
    return (
        <div className={cls.StatCardContainer}>
            <div
                className={cn(cls.Value, {
                    [cls.info]: level == Level.INFO,
                    [cls.warn]: level == Level.WARN,
                    [cls.error]: level == Level.ERROR,
                    [cls.good]: level == Level.Good,
                })}
            >
                {value}
            </div>
            <div className={cls.label}>{label}</div>
        </div>
    );
}
