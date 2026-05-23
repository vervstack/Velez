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
    borderless?: boolean;
    nopadding?: boolean
    size?: 'fill' | 'default'

    tooltipContent?: string
}

export default function Button(
    {
        children, title, variant = 'secondary', onClick,
        disabled, isDisabled, type = 'button', leftIcon, fullWidth,
        sm, borderless = false, nopadding = false, size = 'default',
        tooltipContent
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
                    [cls.borderless]: borderless,
                    [cls.nopadding]: nopadding,
                    [cls.fill]: size == 'fill'
                },
            )}
            onClick={onClick}
            disabled={isOff}
            type={type}

            data-tooltip-id={tooltipContent && 'root-tooltip'}
            data-tooltip-content={tooltipContent}
        >
            {leftIcon && <span className={cls.LeftIcon}>{leftIcon}</span>}
            {children ?? title}
        </button>
    );
}
