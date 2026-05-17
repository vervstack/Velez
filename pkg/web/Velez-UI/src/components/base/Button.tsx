import cls from '@/components/base/Button.module.css';
import cn from 'classnames';

interface ButtonProps {
    children?: React.ReactNode;
    /** @deprecated use children */
    title?: string;
    variant?: 'primary' | 'secondary' | 'danger' | 'warn' | 'ghost';
    onClick?: () => void;
    disabled?: boolean;
    /** @deprecated use disabled */
    isDisabled?: boolean;
    type?: 'button' | 'submit' | 'reset';
    leftIcon?: React.ReactNode;
    fullWidth?: boolean;
    sm?: boolean;
}

export default function Button({
                                   children,
                                   title,
                                   variant = 'secondary',
                                   onClick,
                                   disabled,
                                   isDisabled,
                                   type = 'button',
                                   leftIcon,
                                   fullWidth,
                                   sm,
                               }: ButtonProps) {
    const isOff = disabled ?? isDisabled;

    return (
        <button
            className={cn(
                cls.ButtonContainer,
                cls[variant],
                {
                    [cls.sm]: sm,
                    [cls.fullWidth]: fullWidth,
                },
            )}
            onClick={onClick}
            disabled={isOff}
            type={type}
        >
            {leftIcon && <span className={cls.LeftIcon}>{leftIcon}</span>}
            {children ?? title}
        </button>
    );
}
