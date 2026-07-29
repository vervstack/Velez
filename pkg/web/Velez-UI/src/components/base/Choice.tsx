import cls from '@/components/base/Choice.module.css';
import cn from 'classnames';

interface ChoiceProps {
    title: string;
    sub?: string;
    active: boolean;
    disabled?: boolean;
    onClick: () => void;
}

export default function Choice({title, sub, active, disabled, onClick}: ChoiceProps) {
    return (
        <button
            type="button"
            className={cn(cls.ChoiceContainer, { [cls.active]: active })}
            disabled={disabled}
            onClick={onClick}
        >
            {active && <span className={cls.Check}>✓</span>}
            <span className={cn(cls.Title, { [cls.activeTitle]: active })}>{title}</span>
            {sub && <span className={cls.Sub}>{sub}</span>}
        </button>
    );
}
