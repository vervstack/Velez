import cls from '@/components/base/Choice.module.css';
import cn from 'classnames';
import type {ReactNode} from 'react';

interface ChoiceProps {
    title: string;
    sub?: string;
    active: boolean;
    icon?: ReactNode;
    isSuggested?: boolean;
    disabled?: boolean;
    onClick: () => void;
}

export default function Choice({title, sub, icon, active, isSuggested, disabled, onClick}: ChoiceProps) {
    return (
        <button
            type="button"
            className={cn(cls.ChoiceContainer, { [cls.active]: active, [cls.suggested]: isSuggested })}
            disabled={disabled}
            onClick={onClick}
        >
            {active && <span className={cls.Check}>✓</span>}
            {icon ? (
                <span className={cls.TitleRow}>
                    <span className={cls.Icon}>{icon}</span>
                    <span className={cn(cls.Title, { [cls.activeTitle]: active })}>{title}</span>
                </span>
            ) : (
                <span className={cn(cls.Title, { [cls.activeTitle]: active })}>{title}</span>
            )}
            {sub && <span className={cls.Sub}>{sub}</span>}
        </button>
    );
}
